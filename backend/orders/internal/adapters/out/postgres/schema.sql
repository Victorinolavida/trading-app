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
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id uuid NOT NULL,
    public_id uuid NOT NULL,
    trader_id uuid NOT NULL,
    security_id uuid NOT NULL,
    direction text NOT NULL,
    order_type text NOT NULL,
    price bigint,
    yield bigint,
    quantity bigint NOT NULL,
    allocation_qty bigint DEFAULT 0 NOT NULL,
    allocation_status text DEFAULT 'none'::text NOT NULL,
    nominal_value bigint NOT NULL,
    fx bigint NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT orders_alloc_consistency_check CHECK ((((allocation_status = 'none'::text) AND (allocation_qty = 0)) OR ((allocation_status = 'partial'::text) AND (allocation_qty > 0) AND (allocation_qty < quantity)) OR ((allocation_status = 'allocated'::text) AND (allocation_qty = quantity)))),
    CONSTRAINT orders_alloc_qty_check CHECK (((allocation_qty >= 0) AND (allocation_qty <= quantity))),
    CONSTRAINT orders_alloc_status_check CHECK ((allocation_status = ANY (ARRAY['none'::text, 'partial'::text, 'allocated'::text]))),
    CONSTRAINT orders_direction_check CHECK ((direction = ANY (ARRAY['bid'::text, 'ask'::text]))),
    CONSTRAINT orders_limit_price_check CHECK (((order_type <> 'limit'::text) OR (price IS NOT NULL))),
    CONSTRAINT orders_order_type_check CHECK ((order_type = ANY (ARRAY['limit'::text, 'market'::text]))),
    CONSTRAINT orders_price_check CHECK (((price IS NULL) OR (price > 0))),
    CONSTRAINT orders_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT orders_status_check CHECK ((status = ANY (ARRAY['active'::text, 'fully_allocated'::text, 'cancelled'::text, 'expired'::text, 'killed'::text])))
);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: orders orders_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_public_id_key UNIQUE (public_id);


--
-- Name: orders_active_book_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX orders_active_book_idx ON public.orders USING btree (security_id, direction, price) WHERE (status = 'active'::text);


--
-- PostgreSQL database dump complete
--


