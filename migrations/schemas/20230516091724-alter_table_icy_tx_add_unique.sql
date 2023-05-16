
-- +migrate Up
alter table icy_transactions add constraint unique_icy_txn_src_dest_category unique (src_employee_id, dest_employee_id, category, amount, txn_time);
-- +migrate Down
alter table icy_transactions drop constraint unique_icy_txn_src_dest_category;
