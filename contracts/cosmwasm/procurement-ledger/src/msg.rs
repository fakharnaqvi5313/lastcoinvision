use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::state::Bid;
// Referenced only by the `#[returns(Procurement)]` attribute below, which
// QueryResponses needs resolvable in non-release builds; unused in others.
#[allow(unused_imports)]
use crate::state::Procurement;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    /// The sender becomes the requester of a new procurement, in Draft status.
    CreateProcurement {
        title: String,
        approver: String,
        payer: String,
    },
    /// The approver moves a Draft procurement to Approved.
    Approve { id: u64 },
    /// The approver opens an Approved procurement for bidding.
    OpenForBids { id: u64 },
    /// Any address may submit (or replace their own) bid while a procurement is
    /// OpenForBids.
    SubmitBid {
        id: u64,
        amount: u128,
        description: String,
    },
    /// The approver awards the procurement to one of the submitted bidders.
    AwardBid { id: u64, bidder: String },
    /// The requester confirms the awarded vendor's delivery.
    ConfirmDelivery { id: u64 },
    /// The payer confirms payment has been made for a Delivered procurement.
    ConfirmPayment { id: u64 },
    /// The requester or approver cancels a procurement that hasn't been awarded yet.
    Cancel { id: u64 },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Procurement)]
    Procurement { id: u64 },
    #[returns(BidsResponse)]
    Bids { id: u64 },
}

#[cw_serde]
pub struct BidsResponse {
    pub bids: Vec<Bid>,
}
