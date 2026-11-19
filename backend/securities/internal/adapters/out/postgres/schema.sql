--
-- PostgreSQL database dump
--


-- Dumped from database version 17.11
-- Dumped by pg_dump version 17.11

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
-- Name: securities; Type: SCHEMA; Schema: -; Owner: -
--



SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: securities; Type: TABLE; Schema: securities; Owner: -
--

CREATE TABLE securities (
    id uuid NOT NULL,
    public_id uuid NOT NULL,
    isin character(12) NOT NULL,
    type character varying(4) NOT NULL,
    serie character varying(16) NOT NULL,
    issue_date date NOT NULL,
    maturity_date date NOT NULL,
    nominal_value numeric(10,6) NOT NULL,
    currency character varying(3) NOT NULL,
    fx numeric(10,6) DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    enable boolean DEFAULT true,
    name character varying(128) DEFAULT ''::character varying NOT NULL,
    CONSTRAINT securities_currency_check CHECK (((currency)::text = ANY ((ARRAY['MXN'::character varying, 'UDI'::character varying])::text[]))),
    CONSTRAINT securities_type_check CHECK (((type)::text = ANY ((ARRAY['M'::character varying, 'S'::character varying])::text[])))
);


--
-- Name: securities securities_pkey; Type: CONSTRAINT; Schema: securities; Owner: -
--

ALTER TABLE ONLY securities
    ADD CONSTRAINT securities_pkey PRIMARY KEY (id);


--
-- Name: securities securities_public_id_key; Type: CONSTRAINT; Schema: securities; Owner: -
--

ALTER TABLE ONLY securities
    ADD CONSTRAINT securities_public_id_key UNIQUE (public_id);


--
-- Name: securities securities_type_serie_key; Type: CONSTRAINT; Schema: securities; Owner: -
--

ALTER TABLE ONLY securities
    ADD CONSTRAINT securities_type_serie_key UNIQUE (type, serie);


--
-- PostgreSQL database dump complete
--


