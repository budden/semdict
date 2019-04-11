create or replace function grantuserprivilege(p_sduserid bigint, p_privilegekindid int) returns void
language plpgsql strict as $$
 BEGIN
  raise exception 'call to a forward declaration stub'; END;
$$;

\echo *** forward_declarations.sql Done
