use cosmwasm_schema::cw_serde;
use cw_storage_plus::{Item, Map};

#[cw_serde]
pub enum Status {
    Draft,
    Approved,
    OpenForBids,
    Awarded,
    Delivered,
    Paid,
    Cancelled,
}

#[cw_serde]
pub struct Procurement {
    pub id: u64,
    pub requester: String,
    pub approver: String,
    pub payer: String,
    pub title: String,
    pub created_at: u64,
    pub status: Status,
    pub awarded_bidder: Option<String>,
}

#[cw_serde]
pub struct Bid {
    pub bidder: String,
    pub amount: u128,
    pub description: String,
    pub submitted_at: u64,
}

pub const NEXT_ID: Item<u64> = Item::new("next_id");
pub const PROCUREMENTS: Map<u64, Procurement> = Map::new("procurements");

/// (procurement_id, bidder) -> Bid
pub const BIDS: Map<(u64, &str), Bid> = Map::new("bids");
