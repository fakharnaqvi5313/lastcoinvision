#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Order, Response, StdResult,
};
use cw2::set_contract_version;
use cw_storage_plus::Bound;

use crate::error::ContractError;
use crate::msg::{BidsResponse, ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{Bid, Procurement, Status, BIDS, NEXT_ID, PROCUREMENTS};

const CONTRACT_NAME: &str = "crates.io:procurement-ledger";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    NEXT_ID.save(deps.storage, &0)?;
    Ok(Response::new().add_attribute("action", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::CreateProcurement {
            title,
            approver,
            payer,
        } => execute_create(deps, env, info, title, approver, payer),
        ExecuteMsg::Approve { id } => execute_approve(deps, info, id),
        ExecuteMsg::OpenForBids { id } => execute_open_for_bids(deps, info, id),
        ExecuteMsg::SubmitBid {
            id,
            amount,
            description,
        } => execute_submit_bid(deps, env, info, id, amount, description),
        ExecuteMsg::AwardBid { id, bidder } => execute_award_bid(deps, info, id, bidder),
        ExecuteMsg::ConfirmDelivery { id } => execute_confirm_delivery(deps, info, id),
        ExecuteMsg::ConfirmPayment { id } => execute_confirm_payment(deps, info, id),
        ExecuteMsg::Cancel { id } => execute_cancel(deps, info, id),
    }
}

fn load_procurement(deps: Deps, id: u64) -> Result<Procurement, ContractError> {
    PROCUREMENTS
        .may_load(deps.storage, id)?
        .ok_or(ContractError::NotFound { id })
}

fn require_status(p: &Procurement, expected: Status) -> Result<(), ContractError> {
    if p.status != expected {
        return Err(ContractError::WrongStatus {
            id: p.id,
            status: p.status.clone(),
        });
    }
    Ok(())
}

fn execute_create(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    title: String,
    approver: String,
    payer: String,
) -> Result<Response, ContractError> {
    if title.is_empty() {
        return Err(ContractError::EmptyTitle {});
    }

    let id = NEXT_ID.load(deps.storage)?;
    let procurement = Procurement {
        id,
        requester: info.sender.to_string(),
        approver,
        payer,
        title,
        created_at: env.block.time.seconds(),
        status: Status::Draft,
        awarded_bidder: None,
    };
    PROCUREMENTS.save(deps.storage, id, &procurement)?;
    NEXT_ID.save(deps.storage, &(id + 1))?;

    Ok(Response::new()
        .add_attribute("action", "create_procurement")
        .add_attribute("id", id.to_string())
        .add_attribute("requester", procurement.requester))
}

fn execute_approve(deps: DepsMut, info: MessageInfo, id: u64) -> Result<Response, ContractError> {
    let mut p = load_procurement(deps.as_ref(), id)?;
    if info.sender.as_str() != p.approver {
        return Err(ContractError::NotApprover {});
    }
    require_status(&p, Status::Draft)?;

    p.status = Status::Approved;
    PROCUREMENTS.save(deps.storage, id, &p)?;

    Ok(Response::new()
        .add_attribute("action", "approve")
        .add_attribute("id", id.to_string()))
}

fn execute_open_for_bids(
    deps: DepsMut,
    info: MessageInfo,
    id: u64,
) -> Result<Response, ContractError> {
    let mut p = load_procurement(deps.as_ref(), id)?;
    if info.sender.as_str() != p.approver {
        return Err(ContractError::NotApprover {});
    }
    require_status(&p, Status::Approved)?;

    p.status = Status::OpenForBids;
    PROCUREMENTS.save(deps.storage, id, &p)?;

    Ok(Response::new()
        .add_attribute("action", "open_for_bids")
        .add_attribute("id", id.to_string()))
}

fn execute_submit_bid(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    id: u64,
    amount: u128,
    description: String,
) -> Result<Response, ContractError> {
    let p = load_procurement(deps.as_ref(), id)?;
    require_status(&p, Status::OpenForBids)?;
    if amount == 0 {
        return Err(ContractError::ZeroBidAmount {});
    }

    let bidder = info.sender.to_string();
    BIDS.save(
        deps.storage,
        (id, bidder.as_str()),
        &Bid {
            bidder: bidder.clone(),
            amount,
            description,
            submitted_at: env.block.time.seconds(),
        },
    )?;

    Ok(Response::new()
        .add_attribute("action", "submit_bid")
        .add_attribute("id", id.to_string())
        .add_attribute("bidder", bidder)
        .add_attribute("amount", amount.to_string()))
}

fn execute_award_bid(
    deps: DepsMut,
    info: MessageInfo,
    id: u64,
    bidder: String,
) -> Result<Response, ContractError> {
    let mut p = load_procurement(deps.as_ref(), id)?;
    if info.sender.as_str() != p.approver {
        return Err(ContractError::NotApprover {});
    }
    require_status(&p, Status::OpenForBids)?;

    if BIDS
        .may_load(deps.storage, (id, bidder.as_str()))?
        .is_none()
    {
        return Err(ContractError::BidNotFound { id, bidder });
    }

    p.status = Status::Awarded;
    p.awarded_bidder = Some(bidder.clone());
    PROCUREMENTS.save(deps.storage, id, &p)?;

    Ok(Response::new()
        .add_attribute("action", "award_bid")
        .add_attribute("id", id.to_string())
        .add_attribute("bidder", bidder))
}

fn execute_confirm_delivery(
    deps: DepsMut,
    info: MessageInfo,
    id: u64,
) -> Result<Response, ContractError> {
    let mut p = load_procurement(deps.as_ref(), id)?;
    if info.sender.as_str() != p.requester {
        return Err(ContractError::NotRequester {});
    }
    require_status(&p, Status::Awarded)?;

    p.status = Status::Delivered;
    PROCUREMENTS.save(deps.storage, id, &p)?;

    Ok(Response::new()
        .add_attribute("action", "confirm_delivery")
        .add_attribute("id", id.to_string()))
}

fn execute_confirm_payment(
    deps: DepsMut,
    info: MessageInfo,
    id: u64,
) -> Result<Response, ContractError> {
    let mut p = load_procurement(deps.as_ref(), id)?;
    if info.sender.as_str() != p.payer {
        return Err(ContractError::NotPayer {});
    }
    require_status(&p, Status::Delivered)?;

    p.status = Status::Paid;
    PROCUREMENTS.save(deps.storage, id, &p)?;

    Ok(Response::new()
        .add_attribute("action", "confirm_payment")
        .add_attribute("id", id.to_string()))
}

fn execute_cancel(deps: DepsMut, info: MessageInfo, id: u64) -> Result<Response, ContractError> {
    let mut p = load_procurement(deps.as_ref(), id)?;
    if info.sender.as_str() != p.requester && info.sender.as_str() != p.approver {
        return Err(ContractError::NotCancelAuthorized {});
    }
    match p.status {
        Status::Draft | Status::Approved | Status::OpenForBids => {}
        _ => {
            return Err(ContractError::WrongStatus {
                id,
                status: p.status,
            })
        }
    }

    p.status = Status::Cancelled;
    PROCUREMENTS.save(deps.storage, id, &p)?;

    Ok(Response::new()
        .add_attribute("action", "cancel")
        .add_attribute("id", id.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Procurement { id } => to_json_binary(&PROCUREMENTS.load(deps.storage, id)?),
        QueryMsg::Bids { id } => to_json_binary(&query_bids(deps, id)?),
    }
}

fn query_bids(deps: Deps, id: u64) -> StdResult<BidsResponse> {
    let start: Option<Bound<&str>> = None;
    let bids: StdResult<Vec<_>> = BIDS
        .prefix(id)
        .range(deps.storage, start, None, Order::Ascending)
        .map(|item| item.map(|(_, bid)| bid))
        .collect();

    Ok(BidsResponse { bids: bids? })
}
