#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Order, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{AccessLogResponse, ConsentStatusResponse, ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{AccessLogEntry, ConsentRecord, ACCESS_LOG, ACCESS_SEQ, CONSENTS};

const CONTRACT_NAME: &str = "crates.io:consent-registry";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

const DEFAULT_LOG_LIMIT: u32 = 30;
const MAX_LOG_LIMIT: u32 = 100;

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
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
        ExecuteMsg::GrantConsent {
            grantee,
            data_category,
            expires_at,
        } => execute_grant_consent(deps, env, info, grantee, data_category, expires_at),
        ExecuteMsg::RevokeConsent {
            grantee,
            data_category,
        } => execute_revoke_consent(deps, env, info, grantee, data_category),
        ExecuteMsg::RecordAccess {
            patient,
            data_category,
        } => execute_record_access(deps, env, info, patient, data_category),
    }
}

fn execute_grant_consent(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    grantee: String,
    data_category: String,
    expires_at: Option<u64>,
) -> Result<Response, ContractError> {
    if data_category.is_empty() {
        return Err(ContractError::EmptyDataCategory {});
    }
    let patient = info.sender.as_str();

    let record = ConsentRecord {
        granted_at: env.block.time.seconds(),
        expires_at,
        revoked: false,
        revoked_at: None,
    };
    CONSENTS.save(deps.storage, (patient, &grantee, &data_category), &record)?;

    Ok(Response::new()
        .add_attribute("action", "grant_consent")
        .add_attribute("patient", patient)
        .add_attribute("grantee", grantee)
        .add_attribute("data_category", data_category))
}

fn execute_revoke_consent(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    grantee: String,
    data_category: String,
) -> Result<Response, ContractError> {
    let patient = info.sender.as_str();
    let key = (patient, grantee.as_str(), data_category.as_str());

    let mut record = CONSENTS
        .may_load(deps.storage, key)?
        .unwrap_or(ConsentRecord {
            granted_at: env.block.time.seconds(),
            expires_at: None,
            revoked: false,
            revoked_at: None,
        });
    record.revoked = true;
    record.revoked_at = Some(env.block.time.seconds());
    CONSENTS.save(deps.storage, key, &record)?;

    Ok(Response::new()
        .add_attribute("action", "revoke_consent")
        .add_attribute("patient", patient)
        .add_attribute("grantee", grantee)
        .add_attribute("data_category", data_category))
}

fn execute_record_access(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    patient: String,
    data_category: String,
) -> Result<Response, ContractError> {
    let accessor = info.sender.to_string();
    let now = env.block.time.seconds();

    let granted = CONSENTS
        .may_load(deps.storage, (patient.as_str(), accessor.as_str(), data_category.as_str()))?
        .map(|record| record.is_active(now))
        .unwrap_or(false);

    let seq = ACCESS_SEQ
        .may_load(deps.storage, patient.as_str())?
        .unwrap_or(0);
    ACCESS_LOG.save(
        deps.storage,
        (patient.as_str(), seq),
        &AccessLogEntry {
            accessor: accessor.clone(),
            data_category: data_category.clone(),
            accessed_at: now,
            granted,
        },
    )?;
    ACCESS_SEQ.save(deps.storage, patient.as_str(), &(seq + 1))?;

    Ok(Response::new()
        .add_attribute("action", "record_access")
        .add_attribute("patient", patient)
        .add_attribute("accessor", accessor)
        .add_attribute("data_category", data_category)
        .add_attribute("granted", granted.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::ConsentStatus {
            patient,
            grantee,
            data_category,
        } => to_json_binary(&query_consent_status(deps, env, patient, grantee, data_category)?),
        QueryMsg::AccessLog {
            patient,
            start_after,
            limit,
        } => to_json_binary(&query_access_log(deps, patient, start_after, limit)?),
    }
}

fn query_consent_status(
    deps: Deps,
    env: Env,
    patient: String,
    grantee: String,
    data_category: String,
) -> StdResult<ConsentStatusResponse> {
    let record = CONSENTS.may_load(deps.storage, (&patient, &grantee, &data_category))?;
    Ok(match record {
        Some(r) => ConsentStatusResponse {
            granted: r.is_active(env.block.time.seconds()),
            expires_at: r.expires_at,
            revoked: r.revoked,
        },
        None => ConsentStatusResponse {
            granted: false,
            expires_at: None,
            revoked: false,
        },
    })
}

fn query_access_log(
    deps: Deps,
    patient: String,
    start_after: Option<u64>,
    limit: Option<u32>,
) -> StdResult<AccessLogResponse> {
    let limit = limit.unwrap_or(DEFAULT_LOG_LIMIT).min(MAX_LOG_LIMIT) as usize;
    let start = start_after.map(cw_storage_plus::Bound::exclusive);

    let entries: StdResult<Vec<_>> = ACCESS_LOG
        .prefix(patient.as_str())
        .range(deps.storage, start, None, Order::Ascending)
        .take(limit)
        .map(|item| item.map(|(_, entry)| entry))
        .collect();

    Ok(AccessLogResponse { entries: entries? })
}
