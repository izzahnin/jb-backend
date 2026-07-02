--
-- PostgreSQL database dump
--


-- Dumped from database version 15.17
-- Dumped by pg_dump version 15.17

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: set_order_number(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_order_number() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.id IS NULL THEN
        NEW.id := nextval(pg_get_serial_sequence('orders', 'id'));
    END IF;

    NEW.order_number := 'ORD-' || lpad(NEW.id::text, 3, '0');
    RETURN NEW;
END;
$$;


--
-- Name: set_trip_number(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_trip_number() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.id IS NULL THEN
        NEW.id := nextval(pg_get_serial_sequence('trips', 'id'));
    END IF;

    NEW.trip_number := 'TRIP-' || lpad(NEW.id::text, 3, '0');
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    user_id integer NOT NULL,
    action character varying(20) NOT NULL,
    table_name character varying(50) NOT NULL,
    record_id integer NOT NULL,
    old_values text,
    new_values text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT audit_logs_action_check CHECK (((action)::text = ANY ((ARRAY['CREATE'::character varying, 'UPDATE'::character varying, 'DELETE'::character varying])::text[])))
);


--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id integer NOT NULL,
    company_name character varying(150) NOT NULL,
    pic_name character varying(100),
    phone character varying(20),
    email character varying(100),
    address text,
    npwp character varying(50),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    created_by integer,
    updated_by integer
);


--
-- Name: customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.customers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;


--
-- Name: drivers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.drivers (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    license_number character varying(50),
    phone character varying(20),
    status character varying(20) DEFAULT 'available'::character varying,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    created_by integer,
    updated_by integer,
    CONSTRAINT drivers_status_check CHECK (((status)::text = ANY ((ARRAY['available'::character varying, 'on_duty'::character varying, 'off'::character varying])::text[])))
);


--
-- Name: drivers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.drivers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: drivers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.drivers_id_seq OWNED BY public.drivers.id;


--
-- Name: locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations (
    id bigint NOT NULL,
    trip_id integer,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
)
PARTITION BY RANGE (created_at);


--
-- Name: locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.locations_id_seq OWNED BY public.locations.id;


--
-- Name: locations_default; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations_default (
    id bigint DEFAULT nextval('public.locations_id_seq'::regclass) NOT NULL,
    trip_id integer,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id integer NOT NULL,
    order_number character varying(50) NOT NULL,
    customer_id integer,
    admin_id integer,
    origin text NOT NULL,
    destination text NOT NULL,
    origin_lat double precision,
    origin_lng double precision,
    dest_lat double precision,
    dest_lng double precision,
    is_active boolean DEFAULT true,
    total_containers integer DEFAULT 1,
    order_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    status character varying(20) DEFAULT 'pending'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    CONSTRAINT orders_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'partial'::character varying, 'completed'::character varying, 'cancelled'::character varying])::text[])))
);


--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;


--
-- Name: trips; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trips (
    id integer NOT NULL,
    order_id integer,
    truck_id integer,
    driver_id integer,
    trip_number character varying(50) NOT NULL,
    container_number character varying(50),
    seal_number character varying(50),
    is_active boolean DEFAULT true,
    status character varying(20) DEFAULT 'pickup'::character varying,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by integer,
    started_by integer,
    completed_by integer,
    CONSTRAINT trips_status_check CHECK (((status)::text = ANY ((ARRAY['pickup'::character varying, 'in_transit'::character varying, 'delivered'::character varying, 'cancelled'::character varying])::text[])))
);


--
-- Name: trips_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.trips_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trips_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.trips_id_seq OWNED BY public.trips.id;


--
-- Name: trucks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trucks (
    id integer NOT NULL,
    plate_number character varying(20) NOT NULL,
    truck_type character varying(50),
    status character varying(20) DEFAULT 'available'::character varying,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    created_by integer,
    updated_by integer,
    CONSTRAINT trucks_status_check CHECK (((status)::text = ANY ((ARRAY['available'::character varying, 'on_duty'::character varying, 'maintenance'::character varying])::text[])))
);


--
-- Name: trucks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.trucks_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trucks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.trucks_id_seq OWNED BY public.trucks.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(50) NOT NULL,
    password_hash text NOT NULL,
    full_name character varying(100),
    role character varying(20) NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['super_admin'::character varying, 'admin_sales'::character varying, 'admin_ops'::character varying, 'demo'::character varying])::text[])))
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: locations_default; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations ATTACH PARTITION public.locations_default DEFAULT;


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);


--
-- Name: drivers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers ALTER COLUMN id SET DEFAULT nextval('public.drivers_id_seq'::regclass);


--
-- Name: locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations ALTER COLUMN id SET DEFAULT nextval('public.locations_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);


--
-- Name: trips id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips ALTER COLUMN id SET DEFAULT nextval('public.trips_id_seq'::regclass);


--
-- Name: trucks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trucks ALTER COLUMN id SET DEFAULT nextval('public.trucks_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.customers VALUES (1, 'PT. Nusantara Logistik', 'Hendra Wijaya', '0411-234567', 'hendra@nusantara.co.id', 'Jl. Urip Sumoharjo No. 12, Makassar', '01.234.567.8-801.000', true, '2026-05-18 05:03:10.310313+00', NULL, 2, NULL);
INSERT INTO public.customers VALUES (2, 'CV. Maju Bersama', 'Siti Rahayu', '0411-345678', 'siti@majubersama.com', 'Jl. Perintis Kemerdekaan KM 10, Makassar', '02.345.678.9-802.000', true, '2026-05-25 05:03:10.310313+00', NULL, 2, NULL);
INSERT INTO public.customers VALUES (3, 'PT. Sulawesi Cargo', 'Rudi Santoso', '0421-456789', 'rudi@sulcargo.id', 'Jl. Bau Massepe No. 45, Parepare', '03.456.789.0-803.000', true, '2026-06-02 05:03:10.310313+00', NULL, 2, NULL);
INSERT INTO public.customers VALUES (4, 'PT. Timur Jaya Ekspres', 'Farhan Hidayat', '0411-567890', 'farhan@timurjaya.co.id', 'Jl. Pettarani No. 78, Makassar', '04.567.890.1-804.000', true, '2026-06-12 05:03:10.310313+00', NULL, 2, NULL);
INSERT INTO public.customers VALUES (5, 'CV. Berlian Abadi', 'Nurhayati', '0411-678901', 'nurhayati@berlianabadi.com', 'Jl. Boulevard Raya Blok A No. 5, Makassar', '05.678.901.2-805.000', true, '2026-06-22 05:03:10.310313+00', NULL, 2, NULL);


--
-- Data for Name: drivers; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.drivers VALUES (1, 'Ahmad Fauzi', 'B12345678901234', '08123456789', 'available', true, '2026-04-03 05:03:10.321925+00', NULL, 1, NULL);
INSERT INTO public.drivers VALUES (2, 'Dedi Kurniawan', 'B23456789012345', '08234567890', 'on_duty', true, '2026-04-18 05:03:10.321925+00', NULL, 1, NULL);
INSERT INTO public.drivers VALUES (3, 'Eko Saputra', 'B34567890123456', '08345678901', 'available', true, '2026-05-03 05:03:10.321925+00', NULL, 1, NULL);
INSERT INTO public.drivers VALUES (4, 'Fajar Ramadan', 'B45678901234567', '08456789012', 'on_duty', true, '2026-05-18 05:03:10.321925+00', NULL, 1, NULL);
INSERT INTO public.drivers VALUES (5, 'Galih Pratama', 'B56789012345678', '08567890123', 'available', true, '2026-06-02 05:03:10.321925+00', NULL, 1, NULL);


--
-- Data for Name: locations_default; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.locations_default VALUES (1, 3, -5.1169, 119.4048, '2026-06-30 06:03:10.352766+00');
INSERT INTO public.locations_default VALUES (2, 3, -5.098, 119.4152, '2026-06-30 07:03:10.352766+00');
INSERT INTO public.locations_default VALUES (3, 3, -5.0823, 119.4231, '2026-06-30 08:03:10.352766+00');
INSERT INTO public.locations_default VALUES (4, 3, -4.9756, 119.4512, '2026-06-30 10:03:10.352766+00');
INSERT INTO public.locations_default VALUES (5, 3, -4.8502, 119.5128, '2026-06-30 12:03:10.352766+00');
INSERT INTO public.locations_default VALUES (6, 3, -4.778, 119.538, '2026-06-30 14:03:10.352766+00');
INSERT INTO public.locations_default VALUES (7, 3, -4.6891, 119.5612, '2026-06-30 16:03:10.352766+00');
INSERT INTO public.locations_default VALUES (8, 3, -4.5678, 119.589, '2026-06-30 18:03:10.352766+00');
INSERT INTO public.locations_default VALUES (9, 3, -4.4234, 119.6012, '2026-06-30 20:03:10.352766+00');
INSERT INTO public.locations_default VALUES (10, 3, -4.3012, 119.6187, '2026-07-01 09:03:10.352766+00');
INSERT INTO public.locations_default VALUES (11, 3, -4.1256, 119.6215, '2026-07-01 19:03:10.352766+00');
INSERT INTO public.locations_default VALUES (12, 3, -4.0135, 119.6236, '2026-07-02 03:03:10.352766+00');


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.orders VALUES (3, 'ORD-003', 3, 2, 'Pelabuhan Soekarno-Hatta, Makassar', 'Gudang PT. Sulawesi Cargo, Parepare', -5.1169, 119.4048, -4.0135, 119.6236, true, 1, '2026-06-29 05:03:10.335353+00', 'partial', '2026-06-29 05:03:10.335353+00', NULL);
INSERT INTO public.orders VALUES (4, 'ORD-004', 4, 2, 'Gudang PT. Timur Jaya Ekspres, Makassar', 'Pelabuhan Makassar', -5.1478, 119.4327, -5.1169, 119.4048, true, 1, '2026-07-01 05:03:10.337451+00', 'pending', '2026-07-01 05:03:10.337451+00', NULL);
INSERT INTO public.orders VALUES (5, 'ORD-005', 5, 2, 'Pelabuhan Makassar', 'Gudang CV. Berlian Abadi, Palopo', -5.1169, 119.4048, -2.9922, 120.1964, true, 1, '2026-07-02 05:03:10.339357+00', 'pending', '2026-07-02 05:03:10.339357+00', NULL);
INSERT INTO public.orders VALUES (1, 'ORD-001', 1, 2, 'Pelabuhan Soekarno-Hatta, Makassar', 'Pelabuhan Tanjung Perak, Surabaya', -5.1169, 119.4048, -7.1986, 112.7346, true, 1, '2026-06-07 05:03:10.324814+00', 'completed', '2026-06-07 05:03:10.324814+00', '2026-06-10 05:03:10.356912+00');
INSERT INTO public.orders VALUES (2, 'ORD-002', 2, 2, 'Gudang CV. Maju Bersama, Makassar', 'Pelabuhan Makassar GÇö Ekspor', -5.14, 119.4327, -5.1169, 119.4048, true, 1, '2026-06-14 05:03:10.332539+00', 'completed', '2026-06-14 05:03:10.332539+00', '2026-06-16 05:03:10.358699+00');


--
-- Data for Name: trips; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.trips VALUES (1, 1, 1, 1, 'TRIP-001', 'CONT-MKS-2506-001', 'SEAL-JB-001', true, 'delivered', '2026-06-08 05:03:10.341355+00', '2026-06-10 05:03:10.341355+00', '2026-06-07 05:03:10.341355+00', 3, 3, 3);
INSERT INTO public.trips VALUES (2, 2, 3, 3, 'TRIP-002', 'CONT-MKS-2506-002', 'SEAL-JB-002', true, 'delivered', '2026-06-15 05:03:10.3472+00', '2026-06-16 05:03:10.3472+00', '2026-06-14 05:03:10.3472+00', 3, 3, 3);
INSERT INTO public.trips VALUES (3, 3, 2, 2, 'TRIP-003', 'CONT-MKS-2506-003', 'SEAL-JB-003', true, 'in_transit', '2026-06-30 05:03:10.349869+00', NULL, '2026-06-29 05:03:10.349869+00', 3, 3, NULL);


--
-- Data for Name: trucks; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.trucks VALUES (1, 'DD 1234 ABC', 'Fuso Box 6-Roda', 'available', true, '2026-04-03 05:03:10.317908+00', NULL, 1, NULL);
INSERT INTO public.trucks VALUES (2, 'DD 5678 BCD', 'Tronton 10-Roda', 'on_duty', true, '2026-04-18 05:03:10.317908+00', NULL, 1, NULL);
INSERT INTO public.trucks VALUES (3, 'DD 9012 CDE', 'Trailer 40-Feet', 'available', true, '2026-05-03 05:03:10.317908+00', NULL, 1, NULL);
INSERT INTO public.trucks VALUES (4, 'DD 3456 DEF', 'CDD Box 4-Roda', 'on_duty', true, '2026-05-18 05:03:10.317908+00', NULL, 1, NULL);
INSERT INTO public.trucks VALUES (5, 'DD 7890 EFG', 'Fuso Bak Terbuka', 'available', true, '2026-06-02 05:03:10.317908+00', NULL, 1, NULL);


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.users VALUES (1, 'superadmin', '$2a$10$HAn26KmSkp19CgNf.6ngjOfzY94iD0Li155X23.OrH8L2ZSAcOyCC', 'Rizky Aditya', 'super_admin', true, '2026-07-02 05:03:09.990563+00');
INSERT INTO public.users VALUES (2, 'admin.sales', '$2a$10$JEHMT/shGWnHw5HhnUB9M.lJK3eyqu8I.tukeQaZ03pezXiPJdjuO', 'Andi Kurniawan', 'admin_sales', true, '2026-07-02 05:03:09.990563+00');
INSERT INTO public.users VALUES (3, 'admin.ops', '$2a$10$kBSsDdRLLTQmCv6BgHbxKup7OmWaaB/kX9wtW4BAMND241PXN.qUO', 'Bayu Prasetyo', 'admin_ops', true, '2026-07-02 05:03:09.990563+00');
INSERT INTO public.users VALUES (4, 'demo', '$2a$10$0RYL2c6Dh7KrY6llwUJrduBOfvgxwaSBJCBCfInUlSWcB.iAo7xnq', 'Demo Viewer', 'demo', true, '2026-07-02 05:03:09.990563+00');


--
-- Name: audit_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.audit_logs_id_seq', 1, false);


--
-- Name: customers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.customers_id_seq', 5, true);


--
-- Name: drivers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.drivers_id_seq', 5, true);


--
-- Name: locations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.locations_id_seq', 12, true);


--
-- Name: orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.orders_id_seq', 5, true);


--
-- Name: trips_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.trips_id_seq', 3, true);


--
-- Name: trucks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.trucks_id_seq', 5, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_id_seq', 4, true);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: drivers drivers_license_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_license_number_key UNIQUE (license_number);


--
-- Name: drivers drivers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_pkey PRIMARY KEY (id);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: trips trips_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_pkey PRIMARY KEY (id);


--
-- Name: trips trips_trip_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_trip_number_key UNIQUE (trip_number);


--
-- Name: trucks trucks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trucks
    ADD CONSTRAINT trucks_pkey PRIMARY KEY (id);


--
-- Name: trucks trucks_plate_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trucks
    ADD CONSTRAINT trucks_plate_number_key UNIQUE (plate_number);


--
-- Name: trips uq_trips_order_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT uq_trips_order_id UNIQUE (order_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at DESC);


--
-- Name: idx_audit_logs_table_record; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_table_record ON public.audit_logs USING btree (table_name, record_id);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_customers_company_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_company_name ON public.customers USING btree (company_name);


--
-- Name: idx_customers_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_is_active ON public.customers USING btree (is_active);


--
-- Name: idx_drivers_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_drivers_status ON public.drivers USING btree (status);


--
-- Name: idx_locations_trip_id_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_locations_trip_id_created_at ON ONLY public.locations USING btree (trip_id, created_at DESC);


--
-- Name: idx_orders_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_is_active ON public.orders USING btree (is_active);


--
-- Name: idx_trips_driver_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trips_driver_id ON public.trips USING btree (driver_id);


--
-- Name: idx_trips_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trips_is_active ON public.trips USING btree (is_active);


--
-- Name: idx_trips_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trips_order_id ON public.trips USING btree (order_id);


--
-- Name: idx_trips_truck_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trips_truck_id ON public.trips USING btree (truck_id);


--
-- Name: idx_trucks_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trucks_created_at ON public.trucks USING btree (created_at DESC);


--
-- Name: idx_trucks_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trucks_is_active ON public.trucks USING btree (is_active);


--
-- Name: idx_trucks_plate_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trucks_plate_number ON public.trucks USING btree (plate_number);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: locations_default_trip_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX locations_default_trip_id_created_at_idx ON public.locations_default USING btree (trip_id, created_at DESC);


--
-- Name: locations_default_trip_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.idx_locations_trip_id_created_at ATTACH PARTITION public.locations_default_trip_id_created_at_idx;


--
-- Name: orders trg_orders_set_order_number; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_orders_set_order_number BEFORE INSERT ON public.orders FOR EACH ROW EXECUTE FUNCTION public.set_order_number();


--
-- Name: trips trg_trips_set_trip_number; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_trips_set_trip_number BEFORE INSERT ON public.trips FOR EACH ROW EXECUTE FUNCTION public.set_trip_number();


--
-- Name: audit_logs audit_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: customers customers_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: customers customers_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: drivers drivers_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: drivers drivers_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: locations locations_trip_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE public.locations
    ADD CONSTRAINT locations_trip_id_fkey FOREIGN KEY (trip_id) REFERENCES public.trips(id);


--
-- Name: orders orders_admin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id);


--
-- Name: orders orders_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- Name: trips trips_completed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_completed_by_fkey FOREIGN KEY (completed_by) REFERENCES public.users(id);


--
-- Name: trips trips_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: trips trips_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: trips trips_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: trips trips_started_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_started_by_fkey FOREIGN KEY (started_by) REFERENCES public.users(id);


--
-- Name: trips trips_truck_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_truck_id_fkey FOREIGN KEY (truck_id) REFERENCES public.trucks(id);


--
-- Name: trucks trucks_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trucks
    ADD CONSTRAINT trucks_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: trucks trucks_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trucks
    ADD CONSTRAINT trucks_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--


