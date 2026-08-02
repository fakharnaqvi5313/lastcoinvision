use cosmwasm_std::testing::{message_info, mock_dependencies, mock_env};
use cosmwasm_std::Addr;

use crate::contract::{execute, instantiate, query};
use crate::msg::{BidsResponse, ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{Procurement, Status};
use crate::ContractError;

fn setup() -> cosmwasm_std::OwnedDeps<
    cosmwasm_std::testing::MockStorage,
    cosmwasm_std::testing::MockApi,
    cosmwasm_std::testing::MockQuerier,
> {
    let mut deps = mock_dependencies();
    instantiate(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked("anyone"), &[]),
        InstantiateMsg {},
    )
    .unwrap();
    deps
}

const REQUESTER: &str = "requester1";
const APPROVER: &str = "approver1";
const PAYER: &str = "payer1";
const BIDDER_A: &str = "bidder_a";
const BIDDER_B: &str = "bidder_b";

fn create_procurement(deps: cosmwasm_std::DepsMut) -> u64 {
    execute(
        deps,
        mock_env(),
        message_info(&Addr::unchecked(REQUESTER), &[]),
        ExecuteMsg::CreateProcurement {
            title: "Office chairs".to_string(),
            approver: APPROVER.to_string(),
            payer: PAYER.to_string(),
        },
    )
    .unwrap();
    0
}

fn get_status(deps: cosmwasm_std::Deps, id: u64) -> Status {
    let p: Procurement =
        cosmwasm_std::from_json(query(deps, mock_env(), QueryMsg::Procurement { id }).unwrap())
            .unwrap();
    p.status
}

#[test]
fn full_lifecycle_happy_path() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());
    assert_eq!(get_status(deps.as_ref(), id), Status::Draft);

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::Approve { id },
    )
    .unwrap();
    assert_eq!(get_status(deps.as_ref(), id), Status::Approved);

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::OpenForBids { id },
    )
    .unwrap();
    assert_eq!(get_status(deps.as_ref(), id), Status::OpenForBids);

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(BIDDER_A), &[]),
        ExecuteMsg::SubmitBid {
            id,
            amount: 1000,
            description: "50 chairs at $20 each".to_string(),
        },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(BIDDER_B), &[]),
        ExecuteMsg::SubmitBid {
            id,
            amount: 900,
            description: "50 chairs, bulk discount".to_string(),
        },
    )
    .unwrap();

    let bids: BidsResponse = cosmwasm_std::from_json(
        query(deps.as_ref(), mock_env(), QueryMsg::Bids { id }).unwrap(),
    )
    .unwrap();
    assert_eq!(bids.bids.len(), 2);

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::AwardBid {
            id,
            bidder: BIDDER_B.to_string(),
        },
    )
    .unwrap();
    assert_eq!(get_status(deps.as_ref(), id), Status::Awarded);

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(REQUESTER), &[]),
        ExecuteMsg::ConfirmDelivery { id },
    )
    .unwrap();
    assert_eq!(get_status(deps.as_ref(), id), Status::Delivered);

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(PAYER), &[]),
        ExecuteMsg::ConfirmPayment { id },
    )
    .unwrap();
    assert_eq!(get_status(deps.as_ref(), id), Status::Paid);

    let p: Procurement = cosmwasm_std::from_json(
        query(deps.as_ref(), mock_env(), QueryMsg::Procurement { id }).unwrap(),
    )
    .unwrap();
    assert_eq!(p.awarded_bidder, Some(BIDDER_B.to_string()));
}

#[test]
fn approve_rejects_non_approver() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(REQUESTER), &[]),
        ExecuteMsg::Approve { id },
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::NotApprover {}));
}

#[test]
fn cannot_open_for_bids_before_approval() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::OpenForBids { id },
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::WrongStatus { .. }));
}

#[test]
fn cannot_submit_bid_outside_bidding_window() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(BIDDER_A), &[]),
        ExecuteMsg::SubmitBid {
            id,
            amount: 100,
            description: "too early".to_string(),
        },
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::WrongStatus { .. }));
}

#[test]
fn cannot_award_nonexistent_bid() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::Approve { id },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::OpenForBids { id },
    )
    .unwrap();

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::AwardBid {
            id,
            bidder: "nobody".to_string(),
        },
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::BidNotFound { .. }));
}

#[test]
fn cancel_allowed_before_award_only() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());

    // cancel from Draft works
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(REQUESTER), &[]),
        ExecuteMsg::Cancel { id },
    )
    .unwrap();
    assert_eq!(get_status(deps.as_ref(), id), Status::Cancelled);

    // a second procurement, cancel after award should fail
    let id2 = {
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked(REQUESTER), &[]),
            ExecuteMsg::CreateProcurement {
                title: "Desks".to_string(),
                approver: APPROVER.to_string(),
                payer: PAYER.to_string(),
            },
        )
        .unwrap();
        1
    };
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::Approve { id: id2 },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::OpenForBids { id: id2 },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(BIDDER_A), &[]),
        ExecuteMsg::SubmitBid {
            id: id2,
            amount: 500,
            description: "desks".to_string(),
        },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::AwardBid {
            id: id2,
            bidder: BIDDER_A.to_string(),
        },
    )
    .unwrap();

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(REQUESTER), &[]),
        ExecuteMsg::Cancel { id: id2 },
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::WrongStatus { .. }));
}

#[test]
fn confirm_delivery_rejects_non_requester() {
    let mut deps = setup();
    let id = create_procurement(deps.as_mut());
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::Approve { id },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::OpenForBids { id },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(BIDDER_A), &[]),
        ExecuteMsg::SubmitBid {
            id,
            amount: 500,
            description: "chairs".to_string(),
        },
    )
    .unwrap();
    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(APPROVER), &[]),
        ExecuteMsg::AwardBid {
            id,
            bidder: BIDDER_A.to_string(),
        },
    )
    .unwrap();

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&Addr::unchecked(BIDDER_A), &[]),
        ExecuteMsg::ConfirmDelivery { id },
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::NotRequester {}));
}
