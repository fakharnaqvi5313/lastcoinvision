use cosmwasm_schema::cw_serde;
use cw_storage_plus::Map;

/// A single consent grant: `patient` allows `grantee` to access records tagged
/// `data_category`, optionally until `expires_at` (unix seconds). `revoked` is set
/// permanently true by `RevokeConsent` rather than deleting the record, so the grant's
/// full history (when it was granted, when it was revoked) stays auditable.
#[cw_serde]
pub struct ConsentRecord {
    pub granted_at: u64,
    pub expires_at: Option<u64>,
    pub revoked: bool,
    pub revoked_at: Option<u64>,
}

impl ConsentRecord {
    pub fn is_active(&self, now: u64) -> bool {
        if self.revoked {
            return false;
        }
        match self.expires_at {
            Some(exp) => now < exp,
            None => true,
        }
    }
}

/// A single, append-only audit log entry recording that `accessor` attempted to access
/// `patient`'s data in `data_category` at `accessed_at`, and whether consent was valid
/// (`granted`) at that moment. Every attempt is logged, granted or not — that's the point of
/// an access log.
#[cw_serde]
pub struct AccessLogEntry {
    pub accessor: String,
    pub data_category: String,
    pub accessed_at: u64,
    pub granted: bool,
}

/// (patient, grantee, data_category) -> ConsentRecord
pub const CONSENTS: Map<(&str, &str, &str), ConsentRecord> = Map::new("consents");

/// (patient, seq) -> AccessLogEntry
pub const ACCESS_LOG: Map<(&str, u64), AccessLogEntry> = Map::new("access_log");

/// patient -> next access log sequence number
pub const ACCESS_SEQ: Map<&str, u64> = Map::new("access_seq");
