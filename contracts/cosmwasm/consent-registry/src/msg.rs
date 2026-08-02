use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::state::AccessLogEntry;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    /// The patient (sender) grants `grantee` access to their records tagged
    /// `data_category`, optionally expiring at `expires_at` (unix seconds).
    /// Re-granting an existing (grantee, data_category) pair overwrites it — including
    /// clearing a prior revocation — with a fresh grant.
    GrantConsent {
        grantee: String,
        data_category: String,
        expires_at: Option<u64>,
    },
    /// The patient (sender) revokes a previously granted consent. The record is kept
    /// (marked revoked, not deleted) so its history remains auditable.
    RevokeConsent {
        grantee: String,
        data_category: String,
    },
    /// Anyone (typically the party attempting to access `patient`'s data) records an
    /// access attempt. Always succeeds and always appends to the audit log — whether
    /// consent was actually valid at the time is reported both as a response attribute
    /// (`granted`) and in the persisted log entry; this contract does not itself gate any
    /// off-chain data release, it only authorizes-and-logs for the caller to act on.
    RecordAccess {
        patient: String,
        data_category: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(ConsentStatusResponse)]
    ConsentStatus {
        patient: String,
        grantee: String,
        data_category: String,
    },
    #[returns(AccessLogResponse)]
    AccessLog {
        patient: String,
        start_after: Option<u64>,
        limit: Option<u32>,
    },
}

#[cw_serde]
pub struct ConsentStatusResponse {
    pub granted: bool,
    pub expires_at: Option<u64>,
    pub revoked: bool,
}

#[cw_serde]
pub struct AccessLogResponse {
    pub entries: Vec<AccessLogEntry>,
}
