--
-- PostgreSQL database dump
--

\restrict tPrdW9PFkDQJl2ypDoX4LDWTycAcRahJA9g4h2OAYzmJZ8X9lzFZJXHUiO95grW

-- Dumped from database version 17.6 (Debian 17.6-0+deb13u1)
-- Dumped by pg_dump version 17.6 (Debian 17.6-0+deb13u1)

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
-- Name: app; Type: SCHEMA; Schema: -; Owner: exile
--

CREATE SCHEMA app;


ALTER SCHEMA app OWNER TO exile;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: instance_actions; Type: TABLE; Schema: public; Owner: exile
--

CREATE TABLE public.instance_actions (
    id integer NOT NULL,
    spawner_id integer NOT NULL,
    instance_id text NOT NULL,
    action text NOT NULL,
    "timestamp" integer NOT NULL,
    status text,
    details text
);


ALTER TABLE public.instance_actions OWNER TO exile;

--
-- Name: instance_actions_id_seq; Type: SEQUENCE; Schema: public; Owner: exile
--

CREATE SEQUENCE public.instance_actions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.instance_actions_id_seq OWNER TO exile;

--
-- Name: instance_actions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: exile
--

ALTER SEQUENCE public.instance_actions_id_seq OWNED BY public.instance_actions.id;


--
-- Name: players; Type: TABLE; Schema: public; Owner: exile
--

CREATE TABLE public.players (
    name character varying(20)
);


ALTER TABLE public.players OWNER TO exile;

--
-- Name: server_config; Type: TABLE; Schema: public; Owner: exile
--

CREATE TABLE public.server_config (
    id integer NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    type text DEFAULT 'string'::text NOT NULL,
    category text DEFAULT 'system'::text NOT NULL,
    description text,
    is_read_only integer DEFAULT 0,
    requires_restart integer DEFAULT 0,
    updated_at integer NOT NULL,
    updated_by text
);


ALTER TABLE public.server_config OWNER TO exile;

--
-- Name: server_config_id_seq; Type: SEQUENCE; Schema: public; Owner: exile
--

CREATE SEQUENCE public.server_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.server_config_id_seq OWNER TO exile;

--
-- Name: server_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: exile
--

ALTER SEQUENCE public.server_config_id_seq OWNED BY public.server_config.id;


--
-- Name: server_versions; Type: TABLE; Schema: public; Owner: exile
--

CREATE TABLE public.server_versions (
    id integer NOT NULL,
    filename text NOT NULL,
    version text,
    comment text,
    uploaded_at integer NOT NULL,
    is_active integer DEFAULT 0
);


ALTER TABLE public.server_versions OWNER TO exile;

--
-- Name: server_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: exile
--

CREATE SEQUENCE public.server_versions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.server_versions_id_seq OWNER TO exile;

--
-- Name: server_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: exile
--

ALTER SEQUENCE public.server_versions_id_seq OWNED BY public.server_versions.id;


--
-- Name: spawners; Type: TABLE; Schema: public; Owner: exile
--

CREATE TABLE public.spawners (
    id integer NOT NULL,
    region text,
    host text NOT NULL,
    port integer NOT NULL,
    max_instances integer NOT NULL,
    current_instances integer NOT NULL,
    status text,
    last_seen integer NOT NULL,
    game_version text
);


ALTER TABLE public.spawners OWNER TO exile;

--
-- Name: spawners_id_seq; Type: SEQUENCE; Schema: public; Owner: exile
--

CREATE SEQUENCE public.spawners_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.spawners_id_seq OWNER TO exile;

--
-- Name: spawners_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: exile
--

ALTER SEQUENCE public.spawners_id_seq OWNED BY public.spawners.id;


--
-- Name: instance_actions id; Type: DEFAULT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.instance_actions ALTER COLUMN id SET DEFAULT nextval('public.instance_actions_id_seq'::regclass);


--
-- Name: server_config id; Type: DEFAULT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.server_config ALTER COLUMN id SET DEFAULT nextval('public.server_config_id_seq'::regclass);


--
-- Name: server_versions id; Type: DEFAULT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.server_versions ALTER COLUMN id SET DEFAULT nextval('public.server_versions_id_seq'::regclass);


--
-- Name: spawners id; Type: DEFAULT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.spawners ALTER COLUMN id SET DEFAULT nextval('public.spawners_id_seq'::regclass);


--
-- Data for Name: instance_actions; Type: TABLE DATA; Schema: public; Owner: exile
--

COPY public.instance_actions (id, spawner_id, instance_id, action, "timestamp", status, details) FROM stdin;
10	10	EU18-7777	stop	1765767634	success	
11	10	EU18-7777	start	1765767639	success	
12	10	EU18-7777	stop	1765767669	success	
13	10	EU18-7777	start	1765767679	success	
14	10	EU18-7777	stop	1765767760	success	
15	10	EU18-7778	spawn	1765767793	success	
16	10	EU18-7779	spawn	1765767797	success	
17	10	EU18-7780	spawn	1765767801	success	
18	10	EU18-7781	spawn	1765767803	success	
19	10	EU18-7782	spawn	1765767806	success	
20	10	EU18-7783	spawn	1765767809	success	
21	10	EU18-7777	start	1765767818	success	
22	10	EU18-7784	spawn	1765767829	success	
23	10	EU18-7785	spawn	1765767832	success	
24	10	EU18-7786	spawn	1765767834	success	
25	10	EU18-7787	spawn	1765767837	success	
26	10	EU18-7777	stop	1765767917	success	
27	10	EU18-7778	stop	1765767918	success	
28	10	EU18-7779	stop	1765767918	success	
29	10	EU18-7780	stop	1765767918	success	
30	10	EU18-7781	stop	1765767918	success	
31	10	EU18-7782	stop	1765767919	success	
32	10	EU18-7783	stop	1765767919	success	
33	10	EU18-7784	stop	1765767919	success	
34	10	EU18-7785	stop	1765767919	success	
35	10	EU18-7786	stop	1765767920	success	
36	10	EU18-7787	stop	1765767920	success	
37	10	EU18-7777	start	1765767928	success	
38	10	EU18-7778	start	1765767929	success	
39	10	EU18-7779	start	1765767929	success	
40	10	EU18-7780	start	1765767929	success	
41	10	EU18-7781	start	1765767929	success	
42	10	EU18-7782	start	1765767930	success	
43	10	EU18-7783	start	1765767930	success	
44	10	EU18-7784	start	1765767930	success	
45	10	EU18-7785	start	1765767930	success	
46	10	EU18-7786	start	1765767930	success	
47	10	EU18-7787	start	1765767930	success	
48	10	EU18-7777	stop	1765768032	success	
49	10	EU18-7778	stop	1765768033	success	
50	10	EU18-7780	stop	1765768034	success	
51	10	EU18-7783	stop	1765768034	success	
52	10	EU18-7784	stop	1765768034	success	
53	10	EU18-7785	stop	1765768034	success	
54	10	EU18-7779	stop	1765768034	success	
55	10	EU18-7781	stop	1765768034	success	
56	10	EU18-7782	stop	1765768035	success	
57	10	EU18-7786	stop	1765768035	success	
58	10	EU18-7787	stop	1765768035	success	
59	10	EU18-7780	delete	1765768038	success	
60	10	EU18-7779	delete	1765768042	success	
61	10	EU18-7781	delete	1765768046	success	
62	10	EU18-7777	delete	1765768050	success	
63	10	EU18-7783	delete	1765768053	success	
64	10	EU18-7782	delete	1765768057	success	
65	10	EU18-7787	start	1765810497	success	
66	10	EU18-7778	start	1765810497	success	
67	10	EU18-7784	start	1765810497	success	
68	10	EU18-7785	start	1765810497	success	
69	10	EU18-7786	start	1765810497	success	
70	10	EU18-7787	stop	1765840565	success	
71	10	EU18-7778	stop	1765840565	success	
72	10	EU18-7784	stop	1765840565	success	
73	10	EU18-7785	stop	1765840565	success	
74	10	EU18-7786	stop	1765840565	success	
\.


--
-- Data for Name: players; Type: TABLE DATA; Schema: public; Owner: exile
--

COPY public.players (name) FROM stdin;
\.


--
-- Data for Name: server_config; Type: TABLE DATA; Schema: public; Owner: exile
--

COPY public.server_config (id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by) FROM stdin;
1	server_port	8081	int	system	Port for the master server to listen on	0	1	1765761701	system
2	server_ttl	60s	duration	system	How long a spawner is considered alive since last heartbeat	0	0	1765761701	system
3	cleanup_interval	30s	duration	system	Frequency of cleanup routine for inactive spawners	0	0	1765761701	system
4	max_body_size	1MB	string	system	Maximum size for request bodies	0	1	1765761701	system
5	session_timeout	24h	duration	security	Session timeout for dashboard authentication	0	0	1765761701	system
6	max_instances_per_spawner	20	int	spawner	Default maximum instances per spawner	0	0	1765761701	system
\.


--
-- Data for Name: server_versions; Type: TABLE DATA; Schema: public; Owner: exile
--

COPY public.server_versions (id, filename, version, comment, uploaded_at, is_active) FROM stdin;
2	game_server_1765767206.zip	3.0.0	^^^^	1765767206	1
\.


--
-- Data for Name: spawners; Type: TABLE DATA; Schema: public; Owner: exile
--

COPY public.spawners (id, region, host, port, max_instances, current_instances, status, last_seen, game_version) FROM stdin;
10	EU18	localhost	8080	223	5	Online	1765844477	3.0.0
\.


--
-- Name: instance_actions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: exile
--

SELECT pg_catalog.setval('public.instance_actions_id_seq', 74, true);


--
-- Name: server_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: exile
--

SELECT pg_catalog.setval('public.server_config_id_seq', 6, true);


--
-- Name: server_versions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: exile
--

SELECT pg_catalog.setval('public.server_versions_id_seq', 2, true);


--
-- Name: spawners_id_seq; Type: SEQUENCE SET; Schema: public; Owner: exile
--

SELECT pg_catalog.setval('public.spawners_id_seq', 20, true);


--
-- Name: instance_actions instance_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.instance_actions
    ADD CONSTRAINT instance_actions_pkey PRIMARY KEY (id);


--
-- Name: server_config server_config_key_key; Type: CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.server_config
    ADD CONSTRAINT server_config_key_key UNIQUE (key);


--
-- Name: server_config server_config_pkey; Type: CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.server_config
    ADD CONSTRAINT server_config_pkey PRIMARY KEY (id);


--
-- Name: server_versions server_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.server_versions
    ADD CONSTRAINT server_versions_pkey PRIMARY KEY (id);


--
-- Name: spawners spawners_host_port_key; Type: CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.spawners
    ADD CONSTRAINT spawners_host_port_key UNIQUE (host, port);


--
-- Name: spawners spawners_pkey; Type: CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.spawners
    ADD CONSTRAINT spawners_pkey PRIMARY KEY (id);


--
-- Name: idx_spawners_host_port; Type: INDEX; Schema: public; Owner: exile
--

CREATE UNIQUE INDEX idx_spawners_host_port ON public.spawners USING btree (host, port);


--
-- Name: instance_actions instance_actions_spawner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: exile
--

ALTER TABLE ONLY public.instance_actions
    ADD CONSTRAINT instance_actions_spawner_id_fkey FOREIGN KEY (spawner_id) REFERENCES public.spawners(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict tPrdW9PFkDQJl2ypDoX4LDWTycAcRahJA9g4h2OAYzmJZ8X9lzFZJXHUiO95grW

