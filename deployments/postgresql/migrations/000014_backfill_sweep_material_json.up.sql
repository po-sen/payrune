UPDATE address_policy_allocations
   SET sweep_material_json = jsonb_build_object(
         'material_type', 'bitcoin_hd',
         'material_version', 1,
         'chain', chain,
         'network', network,
         'address', address,
         'hd_derivation_path', issuance_ref,
         'account_xpub', address_space_ref,
         'script_type', scheme
       )
 WHERE allocation_status = 'issued'
   AND chain = 'bitcoin'
   AND sweep_material_json IS NULL
   AND COALESCE(network, '') <> ''
   AND COALESCE(address, '') <> ''
   AND COALESCE(issuance_ref, '') <> ''
   AND COALESCE(address_space_ref, '') <> ''
   AND COALESCE(scheme, '') <> '';

WITH receiver_artifact AS (
  SELECT
    '60a060405260405161020f38038061020f8339810160408190526020916056565b6001600160a01b038116604657604051630f62942b60e31b815260040160405180910390fd5b6001600160a01b03166080526081565b5f602082840312156065575f5ffd5b81516001600160a01b0381168114607a575f5ffd5b9392505050565b60805161017161009e5f395f8181605d015260aa01526101715ff3fe60806040526004361061002b575f3560e01c806335faa41614610036578063913e77ad1461004c575f5ffd5b3661003257005b5f5ffd5b348015610041575f5ffd5b5061004a61009b565b005b348015610057575f5ffd5b5061007f7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b03909116815260200160405180910390f35b475f8190036100a75750565b5f7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316826040515f6040518083038185875af1925050503d805f8114610110576040519150601f19603f3d011682016040523d82523d5f602084013e610115565b606091505b5050905080610137576040516313dd85ff60e31b815260040160405180910390fd5b505056fea2646970667358221220d23638fe8b2f634a1c8a9f98dd3fa5b1b5c5319a17f1d57d30d082e7e7d96a9864736f6c63430008220033'::text AS creation_code_hex
),
parsed AS (
  SELECT
    id,
    lower(substring(address_space_ref from 'factory=(0x[0-9A-Fa-f]{40})')) AS factory_address,
    lower(substring(address_space_ref from 'collector=(0x[0-9A-Fa-f]{40})')) AS collector_address,
    lower(substring(address_space_ref from 'init_code_hash=(0x[0-9A-Fa-f]{64})')) AS init_code_hash
  FROM address_policy_allocations
  WHERE allocation_status = 'issued'
    AND chain = 'ethereum'
    AND scheme = 'create2'
    AND sweep_material_json IS NULL
)
UPDATE address_policy_allocations AS a
   SET sweep_material_json = jsonb_build_object(
         'material_type', 'ethereum_create2',
         'material_version', 1,
         'chain', a.chain,
         'network', a.network,
         'address', a.address,
         'predicted_address', a.address,
         'factory_address', parsed.factory_address,
         'collector_address', parsed.collector_address,
         'create2_salt', a.issuance_ref,
         'init_code_hex', '0x' || receiver_artifact.creation_code_hex || lpad(substr(parsed.collector_address, 3), 64, '0'),
         'init_code_hash', parsed.init_code_hash
       )
  FROM parsed, receiver_artifact
 WHERE a.id = parsed.id
   AND COALESCE(a.network, '') <> ''
   AND COALESCE(a.address, '') <> ''
   AND COALESCE(a.issuance_ref, '') <> ''
   AND parsed.factory_address IS NOT NULL
   AND parsed.collector_address IS NOT NULL
   AND parsed.init_code_hash IS NOT NULL;
