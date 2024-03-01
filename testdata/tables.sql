--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2
-- Dumped by pg_dump version 16.2

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
-- Name: task_kind; Type: TYPE; Schema: public; Owner: gevulot
--

CREATE TYPE public.task_kind AS ENUM (
    'proof',
    'verification',
    'pow',
    'nop'
);


ALTER TYPE public.task_kind OWNER TO gevulot;

--
-- Name: task_state; Type: TYPE; Schema: public; Owner: gevulot
--

CREATE TYPE public.task_state AS ENUM (
    'new',
    'pending',
    'running',
    'ready',
    'failed'
);


ALTER TYPE public.task_state OWNER TO gevulot;

--
-- Name: transaction_kind; Type: TYPE; Schema: public; Owner: gevulot
--

CREATE TYPE public.transaction_kind AS ENUM (
    'empty',
    'transfer',
    'stake',
    'unstake',
    'deploy',
    'run',
    'proof',
    'proofkey',
    'verification',
    'cancel'
);


ALTER TYPE public.transaction_kind OWNER TO gevulot;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: acl_whitelist; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.acl_whitelist (
    key character varying(130) NOT NULL
);


ALTER TABLE public.acl_whitelist OWNER TO gevulot;

--
-- Name: deploy; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.deploy (
    tx character varying(64) NOT NULL,
    name character varying(256),
    prover character varying(64),
    verifier character varying(64)
);


ALTER TABLE public.deploy OWNER TO gevulot;

--
-- Name: program; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.program (
    hash character varying(64) NOT NULL,
    name character varying(128) NOT NULL,
    image_file_name character varying(256) NOT NULL,
    image_file_url character varying(1024) NOT NULL,
    image_file_checksum character varying(128) NOT NULL
);


ALTER TABLE public.program OWNER TO gevulot;

--
-- Name: program_input_data; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.program_input_data (
    workflow_step_id bigint NOT NULL,
    file_name character varying(1024) NOT NULL,
    file_url character varying(4096) NOT NULL,
    checksum character varying(512) NOT NULL
);


ALTER TABLE public.program_input_data OWNER TO gevulot;

--
-- Name: program_output_data; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.program_output_data (
    workflow_step_id bigint NOT NULL,
    file_name character varying(1024) NOT NULL,
    source_program character varying(64) NOT NULL
);


ALTER TABLE public.program_output_data OWNER TO gevulot;

--
-- Name: program_resource_requirements; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.program_resource_requirements (
    program_hash character varying(64) NOT NULL,
    memory bigint NOT NULL,
    cpus bigint NOT NULL,
    gpus bigint NOT NULL
);


ALTER TABLE public.program_resource_requirements OWNER TO gevulot;

--
-- Name: proof; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.proof (
    tx character varying(64) NOT NULL,
    parent character varying(64) NOT NULL,
    prover character varying(64),
    proof bytea NOT NULL
);


ALTER TABLE public.proof OWNER TO gevulot;

--
-- Name: proof_key; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.proof_key (
    tx character varying(64) NOT NULL,
    parent character varying(64) NOT NULL,
    key bytea NOT NULL
);


ALTER TABLE public.proof_key OWNER TO gevulot;

--
-- Name: transaction; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.transaction (
    author character varying(130) NOT NULL,
    hash character varying(64) NOT NULL,
    kind public.transaction_kind NOT NULL,
    nonce numeric NOT NULL,
    signature character varying(128) NOT NULL,
    propagated boolean,
    executed boolean
);


ALTER TABLE public.transaction OWNER TO gevulot;

--
-- Name: txfile; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.txfile (
    tx_id character varying(64) NOT NULL,
    name character varying(256) NOT NULL,
    url character varying(2048) NOT NULL,
    checksum character varying(64) NOT NULL
);


ALTER TABLE public.txfile OWNER TO gevulot;

--
-- Name: verification; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.verification (
    tx character varying(64) NOT NULL,
    parent character varying(64) NOT NULL,
    verifier character varying(64),
    verification bytea NOT NULL
);


ALTER TABLE public.verification OWNER TO gevulot;

--
-- Name: workflow_step; Type: TABLE; Schema: public; Owner: gevulot
--

CREATE TABLE public.workflow_step (
    id bigint NOT NULL,
    tx character varying(64) NOT NULL,
    sequence integer NOT NULL,
    program character varying(64) NOT NULL,
    args character varying(512)[]
);


ALTER TABLE public.workflow_step OWNER TO gevulot;

--
-- Name: workflow_step_id_seq; Type: SEQUENCE; Schema: public; Owner: gevulot
--

CREATE SEQUENCE public.workflow_step_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.workflow_step_id_seq OWNER TO gevulot;

--
-- Name: workflow_step_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gevulot
--

ALTER SEQUENCE public.workflow_step_id_seq OWNED BY public.workflow_step.id;


--
-- Name: workflow_step id; Type: DEFAULT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.workflow_step ALTER COLUMN id SET DEFAULT nextval('public.workflow_step_id_seq'::regclass);


--
-- Name: acl_whitelist acl_whitelist_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.acl_whitelist
    ADD CONSTRAINT acl_whitelist_pkey PRIMARY KEY (key);


--
-- Name: deploy deploy_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.deploy
    ADD CONSTRAINT deploy_pkey PRIMARY KEY (tx);


--
-- Name: program_input_data program_input_data_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_input_data
    ADD CONSTRAINT program_input_data_pkey PRIMARY KEY (workflow_step_id, file_name);


--
-- Name: program_output_data program_output_data_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_output_data
    ADD CONSTRAINT program_output_data_pkey PRIMARY KEY (workflow_step_id, file_name);


--
-- Name: program program_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program
    ADD CONSTRAINT program_pkey PRIMARY KEY (hash);


--
-- Name: program_resource_requirements program_resource_requirements_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_resource_requirements
    ADD CONSTRAINT program_resource_requirements_pkey PRIMARY KEY (program_hash);


--
-- Name: proof_key proof_key_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.proof_key
    ADD CONSTRAINT proof_key_pkey PRIMARY KEY (tx);


--
-- Name: proof proof_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.proof
    ADD CONSTRAINT proof_pkey PRIMARY KEY (tx);


--
-- Name: transaction transaction_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (hash);


--
-- Name: verification verification_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.verification
    ADD CONSTRAINT verification_pkey PRIMARY KEY (tx);


--
-- Name: workflow_step workflow_step_id_key; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.workflow_step
    ADD CONSTRAINT workflow_step_id_key UNIQUE (id);


--
-- Name: workflow_step workflow_step_pkey; Type: CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.workflow_step
    ADD CONSTRAINT workflow_step_pkey PRIMARY KEY (tx, sequence);


--
-- Name: workflow_step fk_program; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.workflow_step
    ADD CONSTRAINT fk_program FOREIGN KEY (program) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: program_resource_requirements fk_program_hash; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_resource_requirements
    ADD CONSTRAINT fk_program_hash FOREIGN KEY (program_hash) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: deploy fk_prover; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.deploy
    ADD CONSTRAINT fk_prover FOREIGN KEY (prover) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: proof fk_prover; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.proof
    ADD CONSTRAINT fk_prover FOREIGN KEY (prover) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: program_output_data fk_source_program; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_output_data
    ADD CONSTRAINT fk_source_program FOREIGN KEY (source_program) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: deploy fk_transaction; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.deploy
    ADD CONSTRAINT fk_transaction FOREIGN KEY (tx) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: workflow_step fk_transaction; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.workflow_step
    ADD CONSTRAINT fk_transaction FOREIGN KEY (tx) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: proof_key fk_transaction; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.proof_key
    ADD CONSTRAINT fk_transaction FOREIGN KEY (tx) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: proof fk_transaction1; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.proof
    ADD CONSTRAINT fk_transaction1 FOREIGN KEY (tx) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: verification fk_transaction1; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.verification
    ADD CONSTRAINT fk_transaction1 FOREIGN KEY (tx) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: proof fk_transaction2; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.proof
    ADD CONSTRAINT fk_transaction2 FOREIGN KEY (parent) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: verification fk_transaction2; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.verification
    ADD CONSTRAINT fk_transaction2 FOREIGN KEY (parent) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: txfile fk_tx; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.txfile
    ADD CONSTRAINT fk_tx FOREIGN KEY (tx_id) REFERENCES public.transaction(hash) ON DELETE CASCADE;


--
-- Name: deploy fk_verifier; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.deploy
    ADD CONSTRAINT fk_verifier FOREIGN KEY (verifier) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: verification fk_verifier; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.verification
    ADD CONSTRAINT fk_verifier FOREIGN KEY (verifier) REFERENCES public.program(hash) ON DELETE CASCADE;


--
-- Name: program_input_data fk_workflow_step; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_input_data
    ADD CONSTRAINT fk_workflow_step FOREIGN KEY (workflow_step_id) REFERENCES public.workflow_step(id) ON DELETE CASCADE;


--
-- Name: program_output_data fk_workflow_step; Type: FK CONSTRAINT; Schema: public; Owner: gevulot
--

ALTER TABLE ONLY public.program_output_data
    ADD CONSTRAINT fk_workflow_step FOREIGN KEY (workflow_step_id) REFERENCES public.workflow_step(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

