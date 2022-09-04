-- -------------------------------------------------------------
-- TablePlus 4.8.2(436)
--
-- https://tableplus.com/
--
-- Database: root
-- Generation Time: 2022-09-05 02:12:50.9850
-- -------------------------------------------------------------


-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS exchange_rate_id_seq;

-- Table Definition
CREATE TABLE "public"."exchange_rate" (
    "id" int4 NOT NULL DEFAULT nextval('exchange_rate_id_seq'::regclass),
    "code" varchar,
    "basePrice" float8,
    PRIMARY KEY ("id")
);

INSERT INTO "public"."exchange_rate" ("id", "code", "basePrice") VALUES
(1, 'FRX.KRWUSD', 1363);
