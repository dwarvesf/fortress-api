-- +migrate Up
CREATE TABLE icy_swap_btc_requests (
   id uuid PRIMARY KEY DEFAULT (uuid()),
   profile_id varchar(255) NOT NULL,
   request_code varchar(255) NOT NULL,
   tx_status varchar(255),
   tx_id int,
   btc_address varchar(255),
   timestamp integer,
   amount varchar(255),
   token_name varchar(255),
   token_id varchar(255),
   swap_request_status varchar(255),
   revert_status varchar(255),
   tx_swap text
);

-- +migrate Down
DROP TABLE icy_swap_btc_requests;