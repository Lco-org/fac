--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
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
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: cleanup_expired_otps(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cleanup_expired_otps() RETURNS void
    LANGUAGE plpgsql
    AS $$ -- ┘ä╪º╪¡╪╕ ╪╣┘ä╪º┘à╪º╪¬ ╪º┘ä╪»┘ê┘ä╪º╪▒ ┘ç┘å╪º (╪¿╪»╪º┘è╪⌐ ╪º┘ä┘â┘ê╪»)
BEGIN
    DELETE FROM otp_codes WHERE expires_at < NOW();
END;
$$;


ALTER FUNCTION public.cleanup_expired_otps() OWNER TO postgres;

--
-- Name: generate_transaction_number(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.generate_transaction_number() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

DECLARE

    prefix TEXT;

    seq_num INT;

    year_suffix TEXT;

BEGIN

    year_suffix := TO_CHAR(CURRENT_DATE, 'YY');

    

    -- Determine prefix based on transaction type

    CASE NEW.type

        WHEN 'income_grant' THEN prefix := 'GR';

        WHEN 'income_donation' THEN prefix := 'DN';

        WHEN 'disbursement_custody' THEN prefix := 'CU';

        WHEN 'disbursement_salary' THEN prefix := 'SA';

        WHEN 'disbursement_operational' THEN prefix := 'OP';

        WHEN 'disbursement_donation' THEN prefix := 'DD';

        WHEN 'disbursement_bonus' THEN prefix := 'BN';

        WHEN 'procurement' THEN prefix := 'PR';

        WHEN 'transfer' THEN prefix := 'TR';

        ELSE prefix := 'TX';

    END CASE;

    

    -- Get next sequence number for this year

    SELECT COALESCE(MAX(

        CAST(SUBSTRING(transaction_number FROM 8) AS INT)

    ), 0) + 1 INTO seq_num

    FROM transactions

    WHERE transaction_number LIKE prefix || '-' || year_suffix || '-%';

    

    NEW.transaction_number := prefix || '-' || year_suffix || '-' || LPAD(seq_num::TEXT, 6, '0');

    

    RETURN NEW;

END;

$$;


ALTER FUNCTION public.generate_transaction_number() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    sequence bigint NOT NULL,
    action character varying(50) NOT NULL,
    user_id uuid,
    username character varying(255),
    resource_type character varying(100) NOT NULL,
    resource_id uuid,
    details jsonb DEFAULT '{}'::jsonb,
    ip_address character varying(45) NOT NULL,
    user_agent text,
    previous_hash character varying(64) NOT NULL,
    current_hash character varying(64) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.audit_logs OWNER TO postgres;

--
-- Name: audit_logs_sequence_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.audit_logs_sequence_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.audit_logs_sequence_seq OWNER TO postgres;

--
-- Name: audit_logs_sequence_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.audit_logs_sequence_seq OWNED BY public.audit_logs.sequence;


--
-- Name: dashboard_widgets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_widgets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    widget_key character varying(100) NOT NULL,
    widget_name character varying(255) NOT NULL,
    widget_name_en character varying(255) NOT NULL,
    is_enabled boolean DEFAULT true,
    allowed_roles text[],
    display_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.dashboard_widgets OWNER TO postgres;

--
-- Name: employees; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employees (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    employee_number character varying(50) NOT NULL,
    full_name character varying(255) NOT NULL,
    full_name_ar character varying(255),
    "position" character varying(255),
    position_ar character varying(255),
    department character varying(255),
    department_ar character varying(255),
    phone character varying(50),
    email character varying(255),
    salary double precision,
    join_date timestamp with time zone,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    base_salary numeric(15,2) DEFAULT 0.00,
    monthly_budget numeric(15,2) DEFAULT 0.00,
    salary_currency character varying(3) DEFAULT 'LYD'::character varying,
    CONSTRAINT chk_employees_base_salary CHECK ((base_salary >= (0)::numeric)),
    CONSTRAINT chk_employees_monthly_budget CHECK ((monthly_budget >= (0)::numeric))
);


ALTER TABLE public.employees OWNER TO postgres;

--
-- Name: COLUMN employees.base_salary; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.employees.base_salary IS 'Employee base salary amount';


--
-- Name: COLUMN employees.monthly_budget; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.employees.monthly_budget IS 'Monthly budget allocation for the employee';


--
-- Name: COLUMN employees.salary_currency; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.employees.salary_currency IS 'Currency code for salary (LYD, USD, EUR, GBP)';


--
-- Name: financial_containers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.financial_containers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    name_ar character varying(255),
    type character varying(50) NOT NULL,
    currency character varying(3) NOT NULL,
    balance double precision DEFAULT 0 NOT NULL,
    bank_name character varying(255),
    bank_name_ar character varying(255),
    account_number character varying(255),
    iban character varying(255),
    swift_code character varying(255),
    location character varying(255),
    location_ar character varying(255),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid
);


ALTER TABLE public.financial_containers OWNER TO postgres;

--
-- Name: member_subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.member_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    member_id uuid NOT NULL,
    transaction_id uuid,
    amount numeric(15,2) NOT NULL,
    currency character varying(3) DEFAULT 'LYD'::character varying,
    payment_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    valid_from timestamp with time zone NOT NULL,
    valid_until timestamp with time zone NOT NULL,
    notes text,
    notes_ar text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    CONSTRAINT chk_member_subscriptions_amount CHECK ((amount > (0)::numeric)),
    CONSTRAINT chk_member_subscriptions_validity CHECK ((valid_until > valid_from))
);


ALTER TABLE public.member_subscriptions OWNER TO postgres;

--
-- Name: TABLE member_subscriptions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.member_subscriptions IS 'Tracks member subscription payments with 365-day validity periods';


--
-- Name: members; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.members (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    membership_number character varying(50) NOT NULL,
    full_name character varying(255) NOT NULL,
    full_name_ar character varying(255),
    member_type character varying(50),
    phone character varying(50),
    email character varying(255),
    address text,
    address_ar text,
    join_date timestamp with time zone,
    status character varying(50) DEFAULT 'active'::character varying,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    valid_until timestamp with time zone,
    last_renewal_date timestamp with time zone,
    subscription_amount numeric(15,2) DEFAULT 0.00,
    subscription_currency character varying(3) DEFAULT 'LYD'::character varying
);


ALTER TABLE public.members OWNER TO postgres;

--
-- Name: COLUMN members.valid_until; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.members.valid_until IS 'Date when membership expires (1 year from last deposit/activation)';


--
-- Name: COLUMN members.last_renewal_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.members.last_renewal_date IS 'Date of last subscription renewal/payment';


--
-- Name: COLUMN members.subscription_amount; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.members.subscription_amount IS 'Annual subscription amount';


--
-- Name: COLUMN members.subscription_currency; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.members.subscription_currency IS 'Currency for subscription payments';


--
-- Name: otp_codes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.otp_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    code_hash character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    attempts integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.otp_codes OWNER TO postgres;

--
-- Name: projects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.projects (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(255) NOT NULL,
    name_ar character varying(255),
    description text,
    donor_name character varying(255),
    donor_name_ar character varying(255),
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    budget double precision,
    currency character varying(3) NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    is_locked boolean DEFAULT false,
    locked_at timestamp with time zone,
    lock_reason text,
    locked_by uuid
);


ALTER TABLE public.projects OWNER TO postgres;

--
-- Name: COLUMN projects.is_locked; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.projects.is_locked IS 'Whether the project is locked for new transactions';


--
-- Name: COLUMN projects.locked_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.projects.locked_at IS 'Timestamp when project was locked';


--
-- Name: COLUMN projects.lock_reason; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.projects.lock_reason IS 'Reason for locking the project';


--
-- Name: COLUMN projects.locked_by; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.projects.locked_by IS 'User who locked the project';


--
-- Name: salary_bonuses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.salary_bonuses (
    id uuid NOT NULL,
    employee_id uuid NOT NULL,
    month_year timestamp without time zone NOT NULL,
    amount numeric(15,2) NOT NULL,
    currency character varying(3) NOT NULL,
    reason text NOT NULL,
    reason_ar text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    created_by uuid
);


ALTER TABLE public.salary_bonuses OWNER TO postgres;

--
-- Name: salary_deductions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.salary_deductions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    employee_id uuid NOT NULL,
    month_year date NOT NULL,
    amount numeric(15,2) NOT NULL,
    currency character varying(3) DEFAULT 'LYD'::character varying,
    reason text NOT NULL,
    reason_ar text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    CONSTRAINT chk_salary_deductions_amount CHECK ((amount > (0)::numeric))
);


ALTER TABLE public.salary_deductions OWNER TO postgres;

--
-- Name: TABLE salary_deductions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.salary_deductions IS 'Tracks monthly salary deductions for employees with reasons';


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transactions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    transaction_number character varying(50) NOT NULL,
    type character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    amount double precision NOT NULL,
    currency character varying(3) NOT NULL,
    exchange_rate double precision DEFAULT 1,
    amount_local double precision,
    container_id uuid NOT NULL,
    project_id uuid,
    description text NOT NULL,
    description_ar text,
    payee_payer character varying(255),
    payee_payer_ar character varying(255),
    reference_number character varying(100),
    is_deficit boolean DEFAULT false,
    deficit_approved_by uuid,
    deficit_warning_shown boolean DEFAULT false,
    balance_before double precision,
    balance_after double precision,
    attachment_url text,
    transaction_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid NOT NULL,
    reversed_at timestamp with time zone,
    reversed_by uuid,
    reversal_reason text,
    original_transaction_id uuid,
    employee_id uuid,
    member_id uuid
);


ALTER TABLE public.transactions OWNER TO postgres;

--
-- Name: COLUMN transactions.employee_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.transactions.employee_id IS 'Employee linked to this transaction (mandatory for salary/custody, optional for operational)';


--
-- Name: COLUMN transactions.member_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.transactions.member_id IS 'Member linked to this transaction (for member deposits/contributions)';


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    username character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    full_name character varying(255) NOT NULL,
    full_name_ar character varying(255),
    role character varying(50) NOT NULL,
    token_version integer DEFAULT 1 NOT NULL,
    two_factor_enabled boolean DEFAULT false,
    is_active boolean DEFAULT true,
    last_login timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    permissions text[] DEFAULT '{}'::text[]
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: vw_members_with_status; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_members_with_status AS
 SELECT id,
    membership_number,
    full_name,
    full_name_ar,
    member_type,
    phone,
    email,
    address,
    address_ar,
    join_date,
    status AS original_status,
    valid_until,
    last_renewal_date,
    subscription_amount,
    subscription_currency,
    is_active,
    created_at,
    updated_at,
    created_by,
        CASE
            WHEN (is_active = false) THEN 'inactive'::character varying
            WHEN (valid_until IS NULL) THEN 'pending'::character varying
            WHEN (valid_until < CURRENT_TIMESTAMP) THEN 'frozen'::character varying
            WHEN (valid_until >= CURRENT_TIMESTAMP) THEN 'active'::character varying
            ELSE COALESCE(status, 'pending'::character varying)
        END AS computed_status,
        CASE
            WHEN ((valid_until IS NOT NULL) AND (valid_until >= CURRENT_TIMESTAMP) AND (valid_until < (CURRENT_TIMESTAMP + '30 days'::interval))) THEN true
            ELSE false
        END AS expiring_soon,
        CASE
            WHEN ((valid_until IS NOT NULL) AND (valid_until >= CURRENT_TIMESTAMP)) THEN (EXTRACT(day FROM (valid_until - CURRENT_TIMESTAMP)))::integer
            ELSE 0
        END AS days_remaining
   FROM public.members m;


ALTER VIEW public.vw_members_with_status OWNER TO postgres;

--
-- Name: VIEW vw_members_with_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON VIEW public.vw_members_with_status IS 'View showing members with computed membership status based on validity dates';


--
-- Name: audit_logs sequence; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN sequence SET DEFAULT nextval('public.audit_logs_sequence_seq'::regclass);


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.audit_logs (id, sequence, action, user_id, username, resource_type, resource_id, details, ip_address, user_agent, previous_hash, current_hash, created_at) FROM stdin;
dad0e765-9a95-48a1-ad2c-578db5d6c847	10	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	bb52b932e34678ba8aa1817f9313e4558ee30f62371aa09b49ef27df7a21db9a	080f3170659e6a0bc7bae9fd4ae0b2da53d7737d36478aa4c454e524272d08f7	2025-12-12 01:33:20.658887+02
7feefc65-6b11-42a2-ad93-1cc99f688470	1	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	GENESIS	37f969232876fafd750fc448333f3101c727cff89a3aeea9a5694dd136ad15e8	2025-12-11 19:55:08.059116+02
01dac458-7d75-446e-90f6-dc9391a9fb5e	2	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	37f969232876fafd750fc448333f3101c727cff89a3aeea9a5694dd136ad15e8	601b209228d5542bcf7a9f6118103931db63584d85c29a36f05ccbe13d225388	2025-12-11 20:33:53.032271+02
d567f6da-be7e-4074-b372-b0e8beff2d6d	3	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	601b209228d5542bcf7a9f6118103931db63584d85c29a36f05ccbe13d225388	16b3ad36dcbf0452e041778bec699a604cb82bb12ba9aed2b57c34ff7df70d3d	2025-12-11 20:39:55.218758+02
e4b76a96-9816-473f-9a5b-b3fd60576355	4	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	16b3ad36dcbf0452e041778bec699a604cb82bb12ba9aed2b57c34ff7df70d3d	ad494ad83554f7c9138988214a812904d8ac4f8959e253bed51b2aad2b748d14	2025-12-11 20:45:59.456127+02
1aafd91e-8aa0-4d3d-a085-ea76d7953fc5	5	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	ad494ad83554f7c9138988214a812904d8ac4f8959e253bed51b2aad2b748d14	cb438ab3d658dca14b067653a5e899bd37717671d9444e8c01a012fd3d97390b	2025-12-11 20:51:08.903802+02
7b3aa346-ac1a-4f90-9905-a1ef05ff3a01	6	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	cb438ab3d658dca14b067653a5e899bd37717671d9444e8c01a012fd3d97390b	a9544d30db5e07cdf918ce545842c7b7f9c0218b452e17f89d38127245b633eb	2025-12-12 00:49:23.668278+02
0c880f93-32f1-4aa1-a5b7-178213666c97	7	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	a9544d30db5e07cdf918ce545842c7b7f9c0218b452e17f89d38127245b633eb	30f5e4d475d12edb7b40a44fc7ee15fce4f2aea36164af45e8ff8b7a147b265b	2025-12-12 01:27:00.533993+02
bedb531e-77cc-4329-b28c-313ff6427f1e	8	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	30f5e4d475d12edb7b40a44fc7ee15fce4f2aea36164af45e8ff8b7a147b265b	e73bd43484d72f1f888bb31bbbce29f49700e7aa39d92c132b97bcde9890295f	2025-12-12 01:29:09.735214+02
a6603104-5e06-4e47-b0db-773caa305b72	9	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	e73bd43484d72f1f888bb31bbbce29f49700e7aa39d92c132b97bcde9890295f	bb52b932e34678ba8aa1817f9313e4558ee30f62371aa09b49ef27df7a21db9a	2025-12-12 01:30:26.012455+02
80211134-fc5f-423f-87e7-7f02f99b06bb	11	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	080f3170659e6a0bc7bae9fd4ae0b2da53d7737d36478aa4c454e524272d08f7	34304985d4d8e8bdc902aa92cecafce873c2cc66f87025187d162ef43edb393a	2025-12-12 01:41:29.007232+02
4a95e6bb-82d1-4f4b-9f0f-01ae389ceb29	12	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	34304985d4d8e8bdc902aa92cecafce873c2cc66f87025187d162ef43edb393a	04cbd458fa05daa95700e2cfc69d4cf1fee22c367151ea5e417849d0317629d3	2025-12-12 01:46:21.626823+02
9550dffa-129f-4803-beaa-9453d4dc9d88	13	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.26100.7019	04cbd458fa05daa95700e2cfc69d4cf1fee22c367151ea5e417849d0317629d3	5b4704be92e857b805acff8dfd6eea7bbdd335f80de28332085180f0f6d50880	2025-12-12 01:58:06.695635+02
c73403e2-03c4-4558-a14a-928a4b24ac6c	14	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.26100.7019	5b4704be92e857b805acff8dfd6eea7bbdd335f80de28332085180f0f6d50880	eb4066a9e5047a22b019677429908edf54aef699f98af5cc6da2cf39acdfb371	2025-12-12 01:58:53.000886+02
ee6ad711-a164-4b2d-94fe-e57930c586df	15	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	eb4066a9e5047a22b019677429908edf54aef699f98af5cc6da2cf39acdfb371	d70f6243c04965c77e36d53f38765386517ddfae95b09a6c016881ce801aeb98	2025-12-12 02:09:57.04599+02
43a02247-8292-4cba-ab6d-b35f679770eb	16	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	d70f6243c04965c77e36d53f38765386517ddfae95b09a6c016881ce801aeb98	fb3e8d51928e8a8fb6a85391feea2d849cac587796b09560afe7661ab71d0062	2025-12-12 02:11:54.088171+02
d75fa68f-0d8f-462e-bcc3-c9b0b7bbc60e	17	REVOKE_ALL_SESSIONS	f925b4f4-b2a5-44a6-8422-a04a506ad450	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"action": "All sessions revoked"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	fb3e8d51928e8a8fb6a85391feea2d849cac587796b09560afe7661ab71d0062	5a8ced43ce19abb731745d03d75d4a964fc5e97cbe7884476a0318b7351677ab	2025-12-12 02:11:58.553094+02
fbd5a342-42a4-4b95-a9cc-2d960abf29a2	18	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	5a8ced43ce19abb731745d03d75d4a964fc5e97cbe7884476a0318b7351677ab	02e1236bc905edd60b1da44c4dc9cf23455d7b9951c6b246012bc4b441565f17	2025-12-12 02:12:00.92321+02
6763fae9-6558-4a0c-9c70-09527c38a4a3	19	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	02e1236bc905edd60b1da44c4dc9cf23455d7b9951c6b246012bc4b441565f17	678639158ddeca76c10558e53b16abb5f0b21d14e29dcd47b73fa7581aca0aec	2025-12-12 02:22:50.243766+02
95e5b619-8394-4af5-ad63-b42eaa0d89e7	20	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	678639158ddeca76c10558e53b16abb5f0b21d14e29dcd47b73fa7581aca0aec	bb934e2d8eade67aaea7682980a1ba850ca9f8933c67562c07122ea861d1a6c7	2025-12-12 02:42:09.735962+02
6227b2d8-349c-480d-aeaf-f9baa1fdaed3	21	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	bb934e2d8eade67aaea7682980a1ba850ca9f8933c67562c07122ea861d1a6c7	62ddefd9e70893f6491ce816f09ba700ad6dc83eab568645c9730dd671ae2e90	2025-12-12 03:07:46.901879+02
60a896b1-c232-43da-a0fa-dc6697c3c117	22	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	62ddefd9e70893f6491ce816f09ba700ad6dc83eab568645c9730dd671ae2e90	4a57052d67681b1498989660a04eeaa502d2fda465058b52c6efdfe665d08c4d	2025-12-12 05:31:11.911454+02
10d02a8f-4a0c-46bf-ae31-2ccf11a9bbee	23	LOGIN	\N	Ezy	auth	f925b4f4-b2a5-44a6-8422-a04a506ad450	{"role": "admin"}	127.0.0.1	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	4a57052d67681b1498989660a04eeaa502d2fda465058b52c6efdfe665d08c4d	d5049dd3e22aaeaf601f6389b5b579a52fd4800e49a1edd202c622ce110c916d	2025-12-12 06:17:42.286773+02
\.


--
-- Data for Name: dashboard_widgets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_widgets (id, widget_key, widget_name, widget_name_en, is_enabled, allowed_roles, display_order, created_at, updated_at) FROM stdin;
c41f2cdf-c0b7-4e63-9623-80a5b8f31c89	projects	╪º┘ä┘à╪┤╪º╪▒┘è╪╣	Projects	t	{admin,finance,auditor}	4	2025-12-07 03:20:28.283544+02	2025-12-07 03:20:28.283544+02
e98fa164-a05f-427b-ab3c-3f420a2e8d10	recent_income	╪ó╪«╪▒ ╪º┘ä╪Ñ┘è╪▒╪º╪»╪º╪¬	Recent Income	t	{admin,finance,auditor}	5	2025-12-07 03:20:28.283544+02	2025-12-07 03:20:28.283544+02
5b14a886-932d-4dfd-8480-9d493f449aa8	recent_expenses	╪ó╪«╪▒ ╪º┘ä┘à╪╡╪▒┘ê┘ü╪º╪¬	Recent Expenses	t	{admin,finance,auditor}	6	2025-12-07 03:20:28.283544+02	2025-12-07 03:20:28.283544+02
6fa469c7-7d55-4cd7-9404-f83488b20e08	banks	╪º┘ä╪¡╪│╪º╪¿╪º╪¬ ╪º┘ä┘à╪╡╪▒┘ü┘è╪⌐	Bank Accounts	t	{admin,finance,auditor}	2	2025-12-07 03:20:28.283544+02	2025-12-07 03:30:55.032247+02
02e2d644-88ed-4b73-a72c-2eb9505c2d49	financial_summary	╪º┘ä┘à┘ä╪«╪╡ ╪º┘ä┘à╪º┘ä┘è	Financial Summary	t	{admin,finance,auditor}	1	2025-12-07 03:20:28.283544+02	2025-12-07 03:30:55.287395+02
1d56c10f-a39c-42e6-a54a-22740885af0b	expiring_memberships	╪¬╪º╪▒┘è╪« ╪º┘ä╪º╪╣╪╢╪º╪í ╪º┘ä┘à┘å╪¬╪│╪¿┘è┘å	Expiring Memberships	t	{admin,auditor}	8	2025-12-11 03:54:39.136671+02	2025-12-11 14:41:35.824461+02
0e83d2eb-e6ea-499c-a92a-93be941ef5d0	cash_accounts	╪º┘ä╪«╪▓╪º╪ª┘å ╪º┘ä┘å┘é╪»┘è╪⌐	Cash Accounts	t	{admin,finance,auditor}	3	2025-12-07 03:20:28.283544+02	2025-12-11 14:41:53.005341+02
\.


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.employees (id, employee_number, full_name, full_name_ar, "position", position_ar, department, department_ar, phone, email, salary, join_date, is_active, created_at, updated_at, created_by, base_salary, monthly_budget, salary_currency) FROM stdin;
081898bd-04d5-4505-985b-f689c89ba9c3	asd	asd	asd	as	asd	asd	as	123123	test@lco.org.ly	\N	\N	t	2025-12-12 02:37:25.615814+02	2025-12-12 02:37:25.615814+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	1000.00	0.00	USD
6ba9b7a2-3ea0-431a-8723-16d3dda7c211	001	┘à╪ñ┘è╪» ┘à╪¡┘à╪» ╪º┘ä╪▓╪¿╪º╪»	\N	╪▒╪ª┘è╪│ ┘à╪¼┘ä╪│ ╪º┘ä╪Ñ╪»╪º╪▒╪⌐	╪▒╪ª┘è╪│ ┘à╪¼┘ä╪│ ╪º┘ä╪Ñ╪»╪º╪▒╪⌐	-	-	0910007701	ezbbad@lco.org.ly	\N	\N	t	2025-12-11 05:38:59.760688+02	2025-12-12 02:37:46.661781+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	0.00	0.00	LYD
\.


--
-- Data for Name: financial_containers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.financial_containers (id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, account_number, iban, swift_code, location, location_ar, is_active, created_at, updated_at, created_by) FROM stdin;
b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	Cash	┘å┘é╪»┘è	cash	LYD	154188142702.46		\N				\N	\N	t	2025-12-11 04:59:52.360672+02	2025-12-12 04:15:31.931081+02	f925b4f4-b2a5-44a6-8422-a04a506ad450
69fd31eb-0208-40b3-a402-e086ac53f7ff	jumhuriya	╪¼┘à┘ç┘ê╪▒┘è╪⌐	bank	LYD	-84850050701.5	┘à╪╡╪▒┘ü ╪º┘ä╪¼┘à┘ç┘ê╪▒┘è╪⌐	\N	001	002	003	\N	\N	t	2025-12-12 03:56:39.168335+02	2025-12-12 04:44:45.934566+02	f925b4f4-b2a5-44a6-8422-a04a506ad450
1cf8a6e3-afce-42e7-8562-0c60626e5851	Cash	┘å┘é╪»┘è	cash	USD	0		\N				\N	\N	t	2025-12-11 05:00:07.255819+02	2025-12-11 05:00:07.255819+02	f925b4f4-b2a5-44a6-8422-a04a506ad450
\.


--
-- Data for Name: member_subscriptions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.member_subscriptions (id, member_id, transaction_id, amount, currency, payment_date, valid_from, valid_until, notes, notes_ar, created_at, updated_at, created_by) FROM stdin;
\.


--
-- Data for Name: members; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.members (id, membership_number, full_name, full_name_ar, member_type, phone, email, address, address_ar, join_date, status, is_active, created_at, updated_at, created_by, valid_until, last_renewal_date, subscription_amount, subscription_currency) FROM stdin;
\.


--
-- Data for Name: otp_codes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.otp_codes (id, user_id, code_hash, expires_at, attempts, created_at) FROM stdin;
\.


--
-- Data for Name: projects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.projects (id, code, name, name_ar, description, donor_name, donor_name_ar, start_date, end_date, budget, currency, is_active, created_at, updated_at, created_by, is_locked, locked_at, lock_reason, locked_by) FROM stdin;
e258f7c0-61d2-4730-a861-afbf1107e4b9	LCO	Libyan Organization for Cooperation	╪º┘ä┘à┘å╪╕┘à╪⌐ ╪º┘ä┘ä┘è╪¿┘è╪⌐ ┘ä┘ä╪¬╪╣╪º┘ê┘å	┘â┘à╪¡┘ü╪╕╪⌐ ╪╣╪º┘à╪⌐ 	\N	\N	2025-10-01 02:00:00+02	\N	0	LYD	t	2025-12-11 05:00:58.055661+02	2025-12-11 14:39:54.042542+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	\N	\N	\N
b6f7705b-7548-4e09-8b67-c2c5a41668ab	LICSF	Libyan International Cybersecurity Forum	╪º┘ä┘à┘å╪¬╪»┘ë ╪º┘ä┘ä┘è╪¿┘è ╪º┘ä╪»┘ê┘ä┘è ┘ä┘ä╪º┘à┘å ╪º┘ä╪│┘è╪¿╪▒╪º┘å┘è	\N	\N	\N	2025-11-01 02:00:00+02	\N	350000	LYD	t	2025-12-11 05:03:01.573443+02	2025-12-11 14:40:09.756285+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	\N	\N	\N
\.


--
-- Data for Name: salary_bonuses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.salary_bonuses (id, employee_id, month_year, amount, currency, reason, reason_ar, created_at, updated_at, created_by) FROM stdin;
cc3d99d7-11ae-459d-8063-d8bac6dc040c	6ba9b7a2-3ea0-431a-8723-16d3dda7c211	2025-12-01 00:00:00	100.00	LYD	0		2025-12-12 02:02:56.534651	2025-12-12 02:02:56.534651	f925b4f4-b2a5-44a6-8422-a04a506ad450
93345bf9-aee0-49fc-b386-1aaf90238f85	081898bd-04d5-4505-985b-f689c89ba9c3	2025-12-01 00:00:00	1121.00	USD	111		2025-12-12 02:38:37.874203	2025-12-12 02:38:37.874203	f925b4f4-b2a5-44a6-8422-a04a506ad450
\.


--
-- Data for Name: salary_deductions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.salary_deductions (id, employee_id, month_year, amount, currency, reason, reason_ar, created_at, updated_at, created_by) FROM stdin;
4db43a17-62e7-4b8f-9e98-f8aa0d740806	6ba9b7a2-3ea0-431a-8723-16d3dda7c211	2025-12-01	10.00	LYD	0		2025-12-12 02:02:40.1613+02	2025-12-12 02:02:40.1613+02	f925b4f4-b2a5-44a6-8422-a04a506ad450
1ca4d4cc-5cc9-4668-aa07-9b0d5c666e74	081898bd-04d5-4505-985b-f689c89ba9c3	2025-12-01	11.00	USD	11		2025-12-12 02:38:22.636071+02	2025-12-12 02:38:22.636071+02	f925b4f4-b2a5-44a6-8422-a04a506ad450
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.transactions (id, transaction_number, type, status, amount, currency, exchange_rate, amount_local, container_id, project_id, description, description_ar, payee_payer, payee_payer_ar, reference_number, is_deficit, deficit_approved_by, deficit_warning_shown, balance_before, balance_after, attachment_url, transaction_date, created_at, updated_at, created_by, reversed_at, reversed_by, reversal_reason, original_transaction_id, employee_id, member_id) FROM stdin;
0bfc58d2-a7d9-4dc0-93e5-af1108ca0329	GR-25-000001	income_grant	completed	1000	LYD	1	1000	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪¬╪¼╪▒┘è╪¿┘è╪⌐	╪¬╪¼╪▒┘è╪¿┘è╪⌐	100	\N	00	f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	0	1000	\N	2025-12-11 15:56:27.136469+02	2025-12-11 15:56:27.136+02	2025-12-11 15:56:27.136+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
20036e51-d4ff-4882-9f7e-8faf4804136e	SA-25-000001	disbursement_salary	completed	2110	LYD	1	2110	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	0001	0001	2100 ╪»┘è┘å╪º╪▒	\N		t	f925b4f4-b2a5-44a6-8422-a04a506ad450	t	1000	-1110	\N	2025-12-12 02:43:25.667716+02	2025-12-12 02:43:25.666653+02	2025-12-12 02:43:25.666653+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	081898bd-04d5-4505-985b-f689c89ba9c3	\N
bfbec944-404f-4ac3-ad50-aca73754c0dd	GR-25-000002	income_grant	completed	150000	LYD	1	150000	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	b6f7705b-7548-4e09-8b67-c2c5a41668ab	╪¿╪»╪º┘è╪⌐ ╪«╪▓┘è┘å╪⌐ ┘ä┘ä┘à╪┤╪▒┘ê╪╣ ┘â╪º┘à┘ä╪º┘ï 	╪¿╪»╪º┘è╪⌐ ╪«╪▓┘è┘å╪⌐ ┘ä┘ä┘à╪┤╪▒┘ê╪╣ ┘â╪º┘à┘ä╪º┘ï 	╪┤╪▒┘â╪⌐ ╪╣╪▒╪º╪¿┘è╪º	\N	25121-0001	f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	-1110	148890	\N	2025-12-12 03:26:12.22923+02	2025-12-12 03:26:12.226276+02	2025-12-12 03:26:12.226276+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
777f1294-27bd-42d2-bf26-cd53fd575d27	DN-25-000001	income_donation	completed	500000	LYD	1	500000	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪¬╪¿╪▒╪╣ ┘ä┘ä┘à┘å╪╕┘à╪⌐	╪¬╪¿╪▒╪╣ ┘ä┘ä┘à┘å╪╕┘à╪⌐	┘à╪¼┘ç┘ê┘ä	\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	148890	648890	\N	2025-12-12 03:28:03.09928+02	2025-12-12 03:28:03.098782+02	2025-12-12 03:28:03.098782+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
2c75c774-bf1c-4205-af18-6734d1965c78	DN-25-000002	income_donation	completed	154336842221	LYD	1	154336842221	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	---	---		\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	648890	154337491111	\N	2025-12-12 03:41:18.653004+02	2025-12-12 03:41:18.652359+02	2025-12-12 03:41:18.652359+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
75ee42d0-ca5b-4c7a-9cf2-4f2f683a4dae	DN-25-000003	income_donation	completed	1213123	LYD	1	1213123	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪º┘ä╪¿┘è╪º┘å 	╪º┘ä╪¿┘è╪º┘å 	╪º┘ä┘à╪¬╪¿╪▒╪╣	\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	154337491111	154338704234	\N	2025-12-12 03:46:13.110577+02	2025-12-12 03:46:13.109986+02	2025-12-12 03:46:13.109986+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
b80a578b-fd05-40a8-a216-19a9c4c676d1	CU-25-000001	disbursement_custody	completed	12412	LYD	1	12412	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪º┘ä╪¿┘è╪º┘å *\n	╪º┘ä╪¿┘è╪º┘å *\n	┘à┘ä╪º╪¡╪╕╪º╪¬ ╪Ñ╪╢╪º┘ü┘è╪⌐	\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	154338704234	154338691822	\N	2025-12-12 03:46:41.888454+02	2025-12-12 03:46:41.887769+02	2025-12-12 03:46:41.887769+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	6ba9b7a2-3ea0-431a-8723-16d3dda7c211	\N
4839f716-0554-4011-9798-b7c5e9f54dce	SA-25-000002	disbursement_salary	completed	12111	LYD	1	12111	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪º┘ä╪¿┘è╪º┘å *\n\n.	╪º┘ä╪¿┘è╪º┘å *\n\n.	┘à┘ä╪º╪¡╪╕╪º╪¬ ╪Ñ╪╢╪º┘ü┘è╪⌐	\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	154338691822	154338679711	\N	2025-12-12 03:47:24.342503+02	2025-12-12 03:47:24.342589+02	2025-12-12 03:47:24.342589+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	6ba9b7a2-3ea0-431a-8723-16d3dda7c211	\N
2122fef5-1ec7-454c-952d-e396eec7f312	OP-25-000001	disbursement_operational	completed	48.54	LYD	1	48.54	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪º┘ä╪¿┘è╪º┘å ╪º┘ä╪¿┘è╪º┘å \nasdasd\nkj	╪º┘ä╪¿┘è╪º┘å ╪º┘ä╪¿┘è╪º┘å \nasdasd\nkj	┘à┘ä╪º╪¡╪╕╪º╪¬ ╪Ñ╪╢╪º┘ü┘è╪⌐	\N	╪▒┘é┘à ╪º┘ä┘à╪▒╪¼╪╣	f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	154338679711	154338679662.46	\N	2025-12-12 03:48:30.427669+02	2025-12-12 03:48:30.427233+02	2025-12-12 03:48:30.427233+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
99a36c2d-9a7c-4bb8-9a9a-4c3addc78416	CU-25-000002	disbursement_custody	completed	582000	LYD	1	582000	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	e258f7c0-61d2-4730-a861-afbf1107e4b9	410	410		\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	154338679662.46	154338097662.46	\N	2025-12-12 03:56:08.417005+02	2025-12-12 03:56:08.41583+02	2025-12-12 03:56:08.41583+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	6ba9b7a2-3ea0-431a-8723-16d3dda7c211	\N
267b5f52-d85c-4da9-8c2f-46d4bea53eb1	TRF-OUT-267b5f52	transfer	completed	5000	LYD	1	5000	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	\N	5000	\N	\N	\N		f	\N	f	154338097662.46	154338092662.46	\N	2025-12-12 03:57:40.416848+02	2025-12-12 03:57:40.416201+02	2025-12-12 03:57:40.416201+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
88b13b37-0968-4170-8b2e-4fe22d733576	TRF-IN-88b13b37	transfer	completed	5000	LYD	1	5000	69fd31eb-0208-40b3-a402-e086ac53f7ff	\N	5000	\N	\N	\N		f	\N	f	0	5000	\N	2025-12-12 03:57:40.416848+02	2025-12-12 03:57:40.416201+02	2025-12-12 03:57:40.416201+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	267b5f52-d85c-4da9-8c2f-46d4bea53eb1	\N	\N
6d858b93-24e8-4f24-859c-5f95ddc50fff	CU-25-000003	disbursement_custody	completed	1450.5	LYD	1	1450.5	69fd31eb-0208-40b3-a402-e086ac53f7ff	e258f7c0-61d2-4730-a861-afbf1107e4b9	╪╣┘ç╪»╪⌐	╪╣┘ç╪»╪⌐		\N		f	f925b4f4-b2a5-44a6-8422-a04a506ad450	f	5000	3549.5	\N	2025-12-12 04:00:05.551265+02	2025-12-12 04:00:05.550306+02	2025-12-12 04:00:05.550306+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	081898bd-04d5-4505-985b-f689c89ba9c3	\N
8321b3d5-9eec-4b73-9dd3-ab8a4b61df85	TRF-OUT-8321b3d5	transfer	completed	150000000	LYD	1	150000000	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	\N	4	\N	\N	\N		f	\N	f	154338092662.46	154188092662.46	\N	2025-12-12 04:09:23.042463+02	2025-12-12 04:09:23.041856+02	2025-12-12 04:09:23.041856+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
0d82f134-1a30-4478-ae7d-ebcd3c26dc8d	TRF-IN-0d82f134	transfer	completed	150000000	LYD	1	150000000	69fd31eb-0208-40b3-a402-e086ac53f7ff	\N	4	\N	\N	\N		f	\N	f	3549.5	150003549.5	\N	2025-12-12 04:09:23.042463+02	2025-12-12 04:09:23.041856+02	2025-12-12 04:09:23.041856+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	8321b3d5-9eec-4b73-9dd3-ab8a4b61df85	\N	\N
b38b44dd-7ce8-4167-895b-4fc900a9e8df	TRF-OUT-b38b44dd	transfer	completed	50040	LYD	1	50040	69fd31eb-0208-40b3-a402-e086ac53f7ff	\N	1	\N	\N	\N		f	\N	f	150003549.5	149953509.5	\N	2025-12-12 04:15:31.93174+02	2025-12-12 04:15:31.931081+02	2025-12-12 04:15:31.931081+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	\N	\N
c3220375-8c2a-495b-a0b7-9b59c7bad0c7	TRF-IN-c3220375	transfer	completed	50040	LYD	1	50040	b912fb49-28c7-4f6d-a9f6-8036f0bb4f67	\N	1	\N	\N	\N		f	\N	f	154188092662.46	154188142702.46	\N	2025-12-12 04:15:31.93174+02	2025-12-12 04:15:31.931081+02	2025-12-12 04:15:31.931081+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	b38b44dd-7ce8-4167-895b-4fc900a9e8df	\N	\N
0061483d-62ca-40a2-a453-857f0daea1c1	CU-25-000004	disbursement_custody	completed	85000004211	LYD	1	85000004211	69fd31eb-0208-40b3-a402-e086ac53f7ff	b6f7705b-7548-4e09-8b67-c2c5a41668ab	4177	4177		\N		t	f925b4f4-b2a5-44a6-8422-a04a506ad450	t	149953509.5	-84850050701.5	\N	2025-12-12 04:44:45.935268+02	2025-12-12 04:44:45.934566+02	2025-12-12 04:44:45.934566+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	\N	\N	\N	\N	6ba9b7a2-3ea0-431a-8723-16d3dda7c211	\N
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, email, password_hash, full_name, full_name_ar, role, token_version, two_factor_enabled, is_active, last_login, created_at, updated_at, created_by, permissions) FROM stdin;
2ff3cbc9-dfc9-48b0-b6cf-e18c0fd3bc17	demo	demo@lco.org.ly	$argon2id$v=19$m=131072,t=4,p=4$8zi1ZEwQxMV6oCL6K7Wspw$f36Bvjo3mUOC4Ps4bgPXQh+bXH9oWVmIJIOlD1gE3Ik	demo	╪¬╪¼╪▒┘è╪¿┘è	admin	1	f	t	2025-12-11 05:57:27.202984+02	2025-12-11 05:54:19.32214+02	2025-12-11 05:54:19.32214+02	f925b4f4-b2a5-44a6-8422-a04a506ad450	{}
f925b4f4-b2a5-44a6-8422-a04a506ad450	Ezy	ezbbad@lco.org.ly	$argon2id$v=19$m=131072,t=4,p=4$M/G/Kp3vsQ8yk+Dl2tMiHg$J6g0kNRbOKQDXGLkSLqtbDtMp9+oMHkyEoMlPkPXBC8	Moaied Ezbbad	┘à╪ñ┘è╪» ╪º┘ä╪▓╪¿╪º╪»	admin	2	f	t	2025-12-12 06:17:42.28117+02	2025-12-07 02:22:18.213838+02	2025-12-12 02:11:58.548741+02	\N	{}
\.


--
-- Name: audit_logs_sequence_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.audit_logs_sequence_seq', 1, false);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: dashboard_widgets dashboard_widgets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_pkey PRIMARY KEY (id);


--
-- Name: dashboard_widgets dashboard_widgets_widget_key_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_widget_key_key UNIQUE (widget_key);


--
-- Name: employees employees_employee_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_employee_number_key UNIQUE (employee_number);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);


--
-- Name: financial_containers financial_containers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.financial_containers
    ADD CONSTRAINT financial_containers_pkey PRIMARY KEY (id);


--
-- Name: member_subscriptions member_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.member_subscriptions
    ADD CONSTRAINT member_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: members members_membership_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.members
    ADD CONSTRAINT members_membership_number_key UNIQUE (membership_number);


--
-- Name: members members_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.members
    ADD CONSTRAINT members_pkey PRIMARY KEY (id);


--
-- Name: otp_codes otp_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.otp_codes
    ADD CONSTRAINT otp_codes_pkey PRIMARY KEY (id);


--
-- Name: projects projects_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_code_key UNIQUE (code);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: salary_bonuses salary_bonuses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.salary_bonuses
    ADD CONSTRAINT salary_bonuses_pkey PRIMARY KEY (id);


--
-- Name: salary_deductions salary_deductions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.salary_deductions
    ADD CONSTRAINT salary_deductions_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_transaction_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_transaction_number_key UNIQUE (transaction_number);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_audit_action_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_action_type ON public.audit_logs USING btree (action, created_at DESC);


--
-- Name: idx_audit_chain; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_chain ON public.audit_logs USING btree (sequence DESC);


--
-- Name: INDEX idx_audit_chain; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON INDEX public.idx_audit_chain IS 'Critical for tamper-proof audit log verification';


--
-- Name: idx_audit_logs_sequence; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_logs_sequence ON public.audit_logs USING btree (sequence);


--
-- Name: idx_audit_user_activity; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_user_activity ON public.audit_logs USING btree (user_id, created_at DESC);


--
-- Name: idx_containers_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_containers_active ON public.financial_containers USING btree (type, is_active) WHERE (is_active = true);


--
-- Name: idx_containers_currency; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_containers_currency ON public.financial_containers USING btree (currency, type) WHERE (is_active = true);


--
-- Name: idx_employees_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_employees_active ON public.employees USING btree (is_active, full_name) WHERE (is_active = true);


--
-- Name: idx_employees_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_employees_number ON public.employees USING btree (employee_number);


--
-- Name: idx_employees_number_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_employees_number_unique ON public.employees USING btree (employee_number);


--
-- Name: idx_member_subscriptions_member_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_member_subscriptions_member_id ON public.member_subscriptions USING btree (member_id);


--
-- Name: idx_member_subscriptions_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_member_subscriptions_transaction_id ON public.member_subscriptions USING btree (transaction_id) WHERE (transaction_id IS NOT NULL);


--
-- Name: idx_member_subscriptions_valid_until; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_member_subscriptions_valid_until ON public.member_subscriptions USING btree (valid_until);


--
-- Name: idx_members_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_members_active ON public.members USING btree (is_active, full_name) WHERE (is_active = true);


--
-- Name: idx_members_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_members_number ON public.members USING btree (membership_number);


--
-- Name: idx_members_number_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_members_number_unique ON public.members USING btree (membership_number);


--
-- Name: idx_otp_expires; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_otp_expires ON public.otp_codes USING btree (expires_at);


--
-- Name: INDEX idx_otp_expires; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON INDEX public.idx_otp_expires IS 'Enables efficient cleanup of expired OTP codes';


--
-- Name: idx_otp_rate_limit; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_otp_rate_limit ON public.otp_codes USING btree (user_id, created_at);


--
-- Name: idx_otp_user_latest; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_otp_user_latest ON public.otp_codes USING btree (user_id, created_at DESC);


--
-- Name: idx_projects_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_active ON public.projects USING btree (is_active, start_date DESC) WHERE (is_active = true);


--
-- Name: idx_projects_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_code ON public.projects USING btree (code);


--
-- Name: idx_projects_code_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_projects_code_unique ON public.projects USING btree (code);


--
-- Name: idx_projects_is_locked; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_projects_is_locked ON public.projects USING btree (is_locked);


--
-- Name: idx_salary_deductions_employee_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_salary_deductions_employee_id ON public.salary_deductions USING btree (employee_id);


--
-- Name: idx_salary_deductions_employee_month; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_salary_deductions_employee_month ON public.salary_deductions USING btree (employee_id, month_year);


--
-- Name: idx_salary_deductions_month_year; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_salary_deductions_month_year ON public.salary_deductions USING btree (month_year);


--
-- Name: idx_transactions_container_history; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_container_history ON public.transactions USING btree (container_id, transaction_date DESC);


--
-- Name: idx_transactions_container_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_container_id ON public.transactions USING btree (container_id);


--
-- Name: idx_transactions_date_range; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_date_range ON public.transactions USING btree (transaction_date DESC, created_at DESC);


--
-- Name: INDEX idx_transactions_date_range; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON INDEX public.idx_transactions_date_range IS 'Primary index for financial reports - covers 80% of queries';


--
-- Name: idx_transactions_employee_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_employee_id ON public.transactions USING btree (employee_id) WHERE (employee_id IS NOT NULL);


--
-- Name: idx_transactions_member_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_member_id ON public.transactions USING btree (member_id) WHERE (member_id IS NOT NULL);


--
-- Name: idx_transactions_project_history; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_project_history ON public.transactions USING btree (project_id, transaction_date DESC);


--
-- Name: idx_transactions_project_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_project_id ON public.transactions USING btree (project_id);


--
-- Name: idx_transactions_reference; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_reference ON public.transactions USING btree (reference_number) WHERE (reference_number IS NOT NULL);


--
-- Name: idx_transactions_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_status ON public.transactions USING btree (status, transaction_date DESC);


--
-- Name: idx_transactions_type_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_type_date ON public.transactions USING btree (type, transaction_date DESC);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_email_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email_lookup ON public.users USING btree (lower((email)::text)) WHERE (is_active = true);


--
-- Name: idx_users_email_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_email_unique ON public.users USING btree (lower((email)::text));


--
-- Name: idx_users_permissions; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_permissions ON public.users USING gin (permissions);


--
-- Name: idx_users_token_version; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_token_version ON public.users USING btree (id, token_version) WHERE (is_active = true);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: idx_users_username_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_username_lookup ON public.users USING btree (username) WHERE (is_active = true);


--
-- Name: idx_users_username_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_username_unique ON public.users USING btree (lower((username)::text));


--
-- Name: transactions set_transaction_number; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER set_transaction_number BEFORE INSERT ON public.transactions FOR EACH ROW WHEN (((new.transaction_number IS NULL) OR ((new.transaction_number)::text = ''::text))) EXECUTE FUNCTION public.generate_transaction_number();


--
-- Name: employees employees_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: financial_containers financial_containers_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.financial_containers
    ADD CONSTRAINT financial_containers_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: member_subscriptions member_subscriptions_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.member_subscriptions
    ADD CONSTRAINT member_subscriptions_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: member_subscriptions member_subscriptions_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.member_subscriptions
    ADD CONSTRAINT member_subscriptions_member_id_fkey FOREIGN KEY (member_id) REFERENCES public.members(id) ON DELETE CASCADE;


--
-- Name: member_subscriptions member_subscriptions_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.member_subscriptions
    ADD CONSTRAINT member_subscriptions_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE SET NULL;


--
-- Name: members members_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.members
    ADD CONSTRAINT members_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: otp_codes otp_codes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.otp_codes
    ADD CONSTRAINT otp_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: projects projects_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: projects projects_locked_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_locked_by_fkey FOREIGN KEY (locked_by) REFERENCES public.users(id);


--
-- Name: salary_bonuses salary_bonuses_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.salary_bonuses
    ADD CONSTRAINT salary_bonuses_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: salary_bonuses salary_bonuses_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.salary_bonuses
    ADD CONSTRAINT salary_bonuses_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(id);


--
-- Name: salary_deductions salary_deductions_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.salary_deductions
    ADD CONSTRAINT salary_deductions_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: salary_deductions salary_deductions_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.salary_deductions
    ADD CONSTRAINT salary_deductions_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_container_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_container_id_fkey FOREIGN KEY (container_id) REFERENCES public.financial_containers(id);


--
-- Name: transactions transactions_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: transactions transactions_deficit_approved_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_deficit_approved_by_fkey FOREIGN KEY (deficit_approved_by) REFERENCES public.users(id);


--
-- Name: transactions transactions_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(id) ON DELETE SET NULL;


--
-- Name: transactions transactions_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_member_id_fkey FOREIGN KEY (member_id) REFERENCES public.members(id) ON DELETE SET NULL;


--
-- Name: transactions transactions_original_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_original_transaction_id_fkey FOREIGN KEY (original_transaction_id) REFERENCES public.transactions(id);


--
-- Name: transactions transactions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: transactions transactions_reversed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_reversed_by_fkey FOREIGN KEY (reversed_by) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

