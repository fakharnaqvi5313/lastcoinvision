use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("procurement {id} not found")]
    NotFound { id: u64 },

    #[error("only the approver may perform this action")]
    NotApprover {},

    #[error("only the requester may perform this action")]
    NotRequester {},

    #[error("only the payer may perform this action")]
    NotPayer {},

    #[error("only the requester or approver may cancel a procurement")]
    NotCancelAuthorized {},

    #[error("procurement {id} is in status {status:?}, which does not allow this action")]
    WrongStatus { id: u64, status: crate::state::Status },

    #[error("no bid from {bidder} was found for procurement {id}")]
    BidNotFound { id: u64, bidder: String },

    #[error("bid amount must be positive")]
    ZeroBidAmount {},

    #[error("title must not be empty")]
    EmptyTitle {},
}
