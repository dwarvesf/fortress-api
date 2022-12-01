
-- +migrate Up
CREATE TABLE IF NOT EXISTS payrolls (
    id                 uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at         TIMESTAMP(6),
    created_at         TIMESTAMP(6) DEFAULT now(),
    updated_at         TIMESTAMP(6) DEFAULT now(),
    employee_id        uuid NOT NULL,
    total              int4,
    due_date           date,
    month              int4,
    year               int4,
    personal_amount    int4,
    contract_amount    int4,
    bonus              int4,
    bonus_explain      json,
    is_paid            bool,
    wise_amount        float4,
    wise_rate          float4,
    wise_fee           float4,
    notes              text,
    commission         int4,
    commission_explain json,
    accounted_amount   int4
);

ALTER TABLE payrolls
    ADD CONSTRAINT payrolls_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);


CREATE TABLE commissions (
    id                      uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at              TIMESTAMP(6),
    created_at              TIMESTAMP(6) DEFAULT now(),
    updated_at              TIMESTAMP(6) DEFAULT now(),
    name                    text,
    monthly_deal_size_from  numeric,
    monthly_deal_size_to    numeric,
    apply_from              date,
    apply_to                date,
    percentage              jsonb
);

CREATE TABLE invoice_items (
    id                uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at        TIMESTAMP(6),
    created_at        TIMESTAMP(6) DEFAULT now(),
    updated_at        TIMESTAMP(6) DEFAULT now(),
    invoice_id        uuid,
    project_member_id uuid,
    total             int4 DEFAULT 0,
    description       text,
    subtotal          int4,
    discount          int4,
    tax               float4,
    type              text,
    is_external       bool
);

ALTER TABLE invoice_items
    ADD CONSTRAINT invoice_itemss_project_member_id_fkey FOREIGN KEY (project_member_id) REFERENCES project_members(id);
ALTER TABLE invoice_items
    ADD CONSTRAINT invoice_items_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES invoices(id);

CREATE TABLE project_commissions (
    id                 uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at         TIMESTAMP(6),
    created_at         TIMESTAMP(6) DEFAULT now(),
    updated_at         TIMESTAMP(6) DEFAULT now(),
    project_id         uuid NOT NULL,
    commission_id      uuid NOT NULL,
    apply_from         date,
    apply_to           date,
    percentage         jsonb
);

ALTER TABLE project_commissions
    ADD CONSTRAINT project_commissions_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE project_commissions
    ADD CONSTRAINT project_commissions_commission_id_fkey FOREIGN KEY (commission_id) REFERENCES commissions(id);

CREATE TABLE project_commission_objects (
    id                    uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at            TIMESTAMP(6),
    created_at            TIMESTAMP(6) DEFAULT now(),
    updated_at            TIMESTAMP(6) DEFAULT now(),
    project_commission_id uuid NOT NULL,
    project_member_id     uuid NOT NULL
);

ALTER TABLE project_commission_objects
    ADD CONSTRAINT project_commission_objects_project_member_id_fkey FOREIGN KEY (project_member_id) REFERENCES project_members(id);
ALTER TABLE project_commission_objects
    ADD CONSTRAINT project_commission_objects_project_commission_id_fkey FOREIGN KEY (project_commission_id) REFERENCES project_commissions(id);

CREATE TABLE project_commission_receivers (
    id                    uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at            TIMESTAMP(6),
    created_at            TIMESTAMP(6) DEFAULT now(),
    updated_at            TIMESTAMP(6) DEFAULT now(),
    project_commission_id uuid NOT NULL,
    employee_id           uuid NOT NULL
);

ALTER TABLE project_commission_receivers
    ADD CONSTRAINT project_commission_receivers_project_commission_id_fkey FOREIGN KEY (project_commission_id) REFERENCES project_commissions(id);
ALTER TABLE project_commission_receivers
    ADD CONSTRAINT project_commission_receivers_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

CREATE TABLE IF NOT EXISTS employee_base_salaries (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6) DEFAULT now(),
    updated_at      TIMESTAMP(6) DEFAULT now(),
    employee_id     uuid NOT NUlL,
    currency_id     uuid NOT NULL,
    start_date      date,
    payroll_batch   int4,
    personal_amount int8,
    contract_amount int8,
    is_active       bool
);

ALTER TABLE employee_base_salaries
    ADD CONSTRAINT employee_base_salaries_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE employee_base_salaries
    ADD CONSTRAINT employee_base_salaries_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies(id);

CREATE TABLE employee_bonus (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6) DEFAULT now(),
    updated_at  TIMESTAMP(6) DEFAULT now(),
    employee_id uuid,
    amount      int4,
    name        text,
    is_active   bool
);

ALTER TABLE employee_bonus
    ADD CONSTRAINT employee_bonus_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

CREATE TABLE employee_commissions (
    id                             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at                     TIMESTAMP(6),
    created_at                     TIMESTAMP(6) DEFAULT now(),
    updated_at                     TIMESTAMP(6) DEFAULT now(),
    project_id                     uuid NOT NULL,
    project_name                   text,
    invoice_id                     uuid NOT NULL,
    invoice_item_id                uuid,
    employee_id                    uuid NOT NULL,
    project_commission_object_id   uuid,
    project_commission_receiver_id uuid NOT NULL,
    percentage                     numeric,
    amount                         numeric NOT NULL DEFAULT 0,
    conversion_rate                numeric DEFAULT 0,
    is_paid                        bool DEFAULT false,
    formula                        text,
    note                           text,
    paid_at                        TIMESTAMP(6)
);

ALTER TABLE employee_commissions
    ADD CONSTRAINT employee_commissions_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE employee_commissions
    ADD CONSTRAINT employee_commissions_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES invoices(id);
ALTER TABLE employee_commissions
    ADD CONSTRAINT employee_commissions_invoice_item_id_fkey FOREIGN KEY (invoice_item_id) REFERENCES invoice_items(id);
ALTER TABLE employee_commissions
    ADD CONSTRAINT employee_commissions_project_commission_object_id_fkey FOREIGN KEY (project_commission_object_id) REFERENCES project_commission_objects(id);
ALTER TABLE employee_commissions
    ADD CONSTRAINT employee_commissions_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);
ALTER TABLE employee_commissions
    ADD CONSTRAINT employee_commissions_project_commission_receiver_id_fkey FOREIGN KEY (project_commission_receiver_id) REFERENCES project_commission_receivers(id);

-- +migrate Down
DROP TABLE IF EXISTS employee_commissions;
DROP TABLE IF EXISTS employee_bonus;
DROP TABLE IF EXISTS employee_base_salaries;
DROP TABLE IF EXISTS project_commission_receivers;
DROP TABLE IF EXISTS project_commission_objects;
DROP TABLE IF EXISTS project_commissions;
DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS commissions;
DROP TABLE IF EXISTS payrolls;
