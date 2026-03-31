UPDATE address_policy_allocations
   SET sweep_material_json = NULL
 WHERE allocation_status = 'issued'
   AND sweep_material_json IS NOT NULL;
