use cosmwasm_std::testing::{message_info, mock_dependencies, mock_env};
use cosmwasm_std::Addr;

use crate::contract::{execute, instantiate, query};
use crate::msg::{AccessLogResponse, ConsentStatusResponse, ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::ContractError;

fn setup() -> cosmwasm_std::OwnedDeps<
    cosmwasm_std::testing::MockStorage,
    cosmwasm_std::testing::MockApi,
    cosmwasm_std::testing::MockQuerier,
> {
    let mut deps = mock_dependencies();
    let info = message_info(&Addr::unchecked("anyone"), &[]);
    instantiate(deps.as_mut(), mock_env(), info, InstantiateMsg {}).unwrap();
    deps
}

#[test]
fn grant_and_query_consent() {
    let mut deps = setup();
    let patient = Addr::unchecked("patient1");
    let clinic = Addr::unchecked("clinic1");

    let info = message_info(&patient, &[]);
    execute(
        deps.as_mut(),
        mock_env(),
        info,
        ExecuteMsg::GrantConsent {
            grantee: clinic.to_string(),
            data_category: "lab_results".to_string(),
            expires_at: None,
        },
    )
    .unwrap();

    let res: ConsentStatusResponse = cosmwasm_std::from_json(
        query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::ConsentStatus {
                patient: patient.to_string(),
                grantee: clinic.to_string(),
                data_category: "lab_results".to_string(),
            },
        )
        .unwrap(),
    )
    .unwrap();

    assert!(res.granted);
    assert!(!res.revoked);
    assert_eq!(res.expires_at, None);
}

#[test]
fn consent_defaults_to_not_granted() {
    let deps = setup();
    let res: ConsentStatusResponse = cosmwasm_std::from_json(
        query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::ConsentStatus {
                patient: "patient1".to_string(),
                grantee: "clinic1".to_string(),
                data_category: "lab_results".to_string(),
            },
        )
        .unwrap(),
    )
    .unwrap();

    assert!(!res.granted);
}

#[test]
fn revoke_consent_makes_it_inactive() {
    let mut deps = setup();
    let patient = Addr::unchecked("patient1");
    let clinic = Addr::unchecked("clinic1");
    let info = message_info(&patient, &[]);

    execute(
        deps.as_mut(),
        mock_env(),
        info.clone(),
        ExecuteMsg::GrantConsent {
            grantee: clinic.to_string(),
            data_category: "lab_results".to_string(),
            expires_at: None,
        },
    )
    .unwrap();

    execute(
        deps.as_mut(),
        mock_env(),
        info,
        ExecuteMsg::RevokeConsent {
            grantee: clinic.to_string(),
            data_category: "lab_results".to_string(),
        },
    )
    .unwrap();

    let res: ConsentStatusResponse = cosmwasm_std::from_json(
        query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::ConsentStatus {
                patient: patient.to_string(),
                grantee: clinic.to_string(),
                data_category: "lab_results".to_string(),
            },
        )
        .unwrap(),
    )
    .unwrap();

    assert!(!res.granted);
    assert!(res.revoked);
}

#[test]
fn expired_consent_is_not_active() {
    let mut deps = setup();
    let patient = Addr::unchecked("patient1");
    let clinic = Addr::unchecked("clinic1");
    let info = message_info(&patient, &[]);
    let env = mock_env();
    let now = env.block.time.seconds();

    execute(
        deps.as_mut(),
        env.clone(),
        info,
        ExecuteMsg::GrantConsent {
            grantee: clinic.to_string(),
            data_category: "lab_results".to_string(),
            expires_at: Some(now), // already expired at query time (now >= expires_at)
        },
    )
    .unwrap();

    let res: ConsentStatusResponse = cosmwasm_std::from_json(
        query(
            deps.as_ref(),
            env,
            QueryMsg::ConsentStatus {
                patient: patient.to_string(),
                grantee: clinic.to_string(),
                data_category: "lab_results".to_string(),
            },
        )
        .unwrap(),
    )
    .unwrap();

    assert!(!res.granted);
    assert!(!res.revoked);
}

#[test]
fn record_access_logs_granted_and_denied_attempts() {
    let mut deps = setup();
    let patient = Addr::unchecked("patient1");
    let clinic = Addr::unchecked("clinic1");
    let stranger = Addr::unchecked("stranger1");

    execute(
        deps.as_mut(),
        mock_env(),
        message_info(&patient, &[]),
        ExecuteMsg::GrantConsent {
            grantee: clinic.to_string(),
            data_category: "lab_results".to_string(),
            expires_at: None,
        },
    )
    .unwrap();

    // clinic has consent: access is granted and logged as such
    let res = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&clinic, &[]),
        ExecuteMsg::RecordAccess {
            patient: patient.to_string(),
            data_category: "lab_results".to_string(),
        },
    )
    .unwrap();
    assert!(res
        .attributes
        .iter()
        .any(|a| a.key == "granted" && a.value == "true"));

    // stranger has no consent: access is still LOGGED but reported as denied, and the
    // call itself does not error — recording an unauthorized attempt is the point.
    let res = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&stranger, &[]),
        ExecuteMsg::RecordAccess {
            patient: patient.to_string(),
            data_category: "lab_results".to_string(),
        },
    )
    .unwrap();
    assert!(res
        .attributes
        .iter()
        .any(|a| a.key == "granted" && a.value == "false"));

    let log: AccessLogResponse = cosmwasm_std::from_json(
        query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::AccessLog {
                patient: patient.to_string(),
                start_after: None,
                limit: None,
            },
        )
        .unwrap(),
    )
    .unwrap();

    assert_eq!(log.entries.len(), 2);
    assert_eq!(log.entries[0].accessor, clinic.to_string());
    assert!(log.entries[0].granted);
    assert_eq!(log.entries[1].accessor, stranger.to_string());
    assert!(!log.entries[1].granted);
}

#[test]
fn grant_consent_rejects_empty_data_category() {
    let mut deps = setup();
    let patient = Addr::unchecked("patient1");

    let err = execute(
        deps.as_mut(),
        mock_env(),
        message_info(&patient, &[]),
        ExecuteMsg::GrantConsent {
            grantee: "clinic1".to_string(),
            data_category: "".to_string(),
            expires_at: None,
        },
    )
    .unwrap_err();

    assert!(matches!(err, ContractError::EmptyDataCategory {}));
}
