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
   swap_request_error text,
   revert_status varchar(255),
   revert_error text,
   withdraw_status varchar(255),
   withdraw_error text,
   tx_swap text,
   tx_deposit text,
   created_at TIMESTAMP(6) DEFAULT NOW(),
   updated_at TIMESTAMP(6) DEFAULT NOW()
);

CREATE INDEX idx_icy_swap_btc_profile_id ON icy_swap_btc_requests(profile_id);
CREATE INDEX idx_icy_swap_btc_request_code ON icy_swap_btc_requests(request_code);

-- +migrate Down
DROP INDEX IF EXISTS idx_icy_swap_btc_profile_id;
DROP INDEX IF EXISTS idx_icy_swap_btc_request_code;
DROP TABLE icy_swap_btc_requests;