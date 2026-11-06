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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: securities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.securities (
    id uuid NOT NULL,
    public_id uuid NOT NULL,
    isin character(12) NOT NULL,
    name character varying(128) NOT NULL DEFAULT '',
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
    CONSTRAINT securities_currency_check CHECK (((currency)::text = ANY ((ARRAY['MXN'::character varying, 'UDI'::character varying])::text[]))),
    CONSTRAINT securities_type_check CHECK (((type)::text = ANY ((ARRAY['M'::character varying, 'S'::character varying])::text[]))),
    CONSTRAINT securities_name_not_empty CHECK (char_length(trim(name)) > 0)
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    password_hash text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_first_name_len CHECK (((char_length(first_name) >= 3) AND (char_length(first_name) <= 49))),
    CONSTRAINT users_last_name_len CHECK (((char_length(last_name) >= 3) AND (char_length(last_name) <= 49)))
);


--
-- Name: securities securities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.securities
    ADD CONSTRAINT securities_pkey PRIMARY KEY (id);


--
-- Name: securities securities_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.securities
    ADD CONSTRAINT securities_public_id_key UNIQUE (public_id);


--
-- Name: securities securities_type_serie_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.securities
    ADD CONSTRAINT securities_type_serie_key UNIQUE (type, serie);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_created_at_idx ON public.users USING btree (created_at);


--
-- PostgreSQL database dump complete
--


