DO $$
BEGIN
  RAISE EXCEPTION
    '000016_drop_issuance_ref_columns is irreversible: issuance_ref columns were dropped; restore from backup or use a forward repair migration';
END;
$$;
