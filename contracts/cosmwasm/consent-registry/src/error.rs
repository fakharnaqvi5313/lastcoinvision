use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("only the patient may manage their own consent records")]
    Unauthorized {},

    #[error("data category must not be empty")]
    EmptyDataCategory {},
}
