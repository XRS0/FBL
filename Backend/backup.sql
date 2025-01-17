--
-- PostgreSQL database dump
--

-- Dumped from database version 15.10 (Debian 15.10-1.pgdg120+1)
-- Dumped by pg_dump version 15.10 (Debian 15.10-1.pgdg120+1)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: match_statistics; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.match_statistics (
    id integer NOT NULL,
    match_id bigint NOT NULL,
    team_id1 bigint NOT NULL,
    team_id2 bigint NOT NULL,
    team1_score integer DEFAULT 0,
    team2_score integer DEFAULT 0,
    created_at timestamp with time zone
);


ALTER TABLE public.match_statistics OWNER TO admin;

--
-- Name: match_statistics_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.match_statistics_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.match_statistics_id_seq OWNER TO admin;

--
-- Name: match_statistics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.match_statistics_id_seq OWNED BY public.match_statistics.id;


--
-- Name: matches; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.matches (
    id integer NOT NULL,
    team1_id bigint NOT NULL,
    team2_id bigint NOT NULL,
    date timestamp without time zone NOT NULL,
    location character varying(255),
    created_at timestamp with time zone
);


ALTER TABLE public.matches OWNER TO admin;

--
-- Name: matches_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.matches_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.matches_id_seq OWNER TO admin;

--
-- Name: matches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.matches_id_seq OWNED BY public.matches.id;


--
-- Name: players; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.players (
    id integer NOT NULL,
    name text NOT NULL,
    height bigint NOT NULL,
    weight bigint NOT NULL,
    "position" text NOT NULL,
    chat_id bigint NOT NULL,
    contact text,
    team_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.players OWNER TO admin;

--
-- Name: players_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.players_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.players_id_seq OWNER TO admin;

--
-- Name: players_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.players_id_seq OWNED BY public.players.id;


--
-- Name: teams; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.teams (
    id integer NOT NULL,
    name text NOT NULL,
    owner_id bigint NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.teams OWNER TO admin;

--
-- Name: teams_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.teams_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.teams_id_seq OWNER TO admin;

--
-- Name: teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.teams_id_seq OWNED BY public.teams.id;


--
-- Name: match_statistics id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.match_statistics ALTER COLUMN id SET DEFAULT nextval('public.match_statistics_id_seq'::regclass);


--
-- Name: matches id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.matches ALTER COLUMN id SET DEFAULT nextval('public.matches_id_seq'::regclass);


--
-- Name: players id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.players ALTER COLUMN id SET DEFAULT nextval('public.players_id_seq'::regclass);


--
-- Name: teams id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.teams ALTER COLUMN id SET DEFAULT nextval('public.teams_id_seq'::regclass);


--
-- Data for Name: match_statistics; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.match_statistics (id, match_id, team_id1, team_id2, team1_score, team2_score, created_at) FROM stdin;
1	1	1	2	0	0	2024-12-29 08:48:04.558151+00
\.


--
-- Data for Name: matches; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.matches (id, team1_id, team2_id, date, location, created_at) FROM stdin;
1	1	2	2024-12-28 20:35:00	ДИВС	2024-12-29 08:47:47.529772+00
\.


--
-- Data for Name: players; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.players (id, name, height, weight, "position", chat_id, contact, team_id, created_at, updated_at) FROM stdin;
1	Максим Александрович Пьянков	182	64	Под кольцом	1324977667	Номер телефона - +79826607484\nTgID - @inpukk	\N	2024-12-29 08:45:58.488362+00	2024-12-29 08:45:58.488362+00
\.


--
-- Data for Name: teams; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.teams (id, name, owner_id, is_active, created_at, updated_at) FROM stdin;
1	Flow	1	t	2024-12-29 08:46:34.13361+00	2024-12-29 08:46:34.13361+00
2	Pupsiki	1	t	2024-12-29 08:47:09.9322+00	2024-12-29 08:47:09.9322+00
\.


--
-- Name: match_statistics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.match_statistics_id_seq', 1, false);


--
-- Name: matches_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.matches_id_seq', 1, true);


--
-- Name: players_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.players_id_seq', 1, true);


--
-- Name: teams_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.teams_id_seq', 2, true);


--
-- Name: match_statistics match_statistics_match_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.match_statistics
    ADD CONSTRAINT match_statistics_match_id_key UNIQUE (match_id);


--
-- Name: match_statistics match_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.match_statistics
    ADD CONSTRAINT match_statistics_pkey PRIMARY KEY (id);


--
-- Name: matches matches_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_pkey PRIMARY KEY (id);


--
-- Name: players players_chat_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_chat_id_key UNIQUE (chat_id);


--
-- Name: players players_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_pkey PRIMARY KEY (id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: idx_players_team_id; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_players_team_id ON public.players USING btree (team_id);


--
-- Name: matches fk_matches_team1; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT fk_matches_team1 FOREIGN KEY (team1_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: matches fk_matches_team2; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT fk_matches_team2 FOREIGN KEY (team2_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: teams fk_teams_owner; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT fk_teams_owner FOREIGN KEY (owner_id) REFERENCES public.players(id);


--
-- Name: players fk_teams_players; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT fk_teams_players FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: match_statistics match_statistics_match_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.match_statistics
    ADD CONSTRAINT match_statistics_match_id_fkey FOREIGN KEY (match_id) REFERENCES public.matches(id) ON DELETE CASCADE;


--
-- Name: match_statistics match_statistics_team_id1_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.match_statistics
    ADD CONSTRAINT match_statistics_team_id1_fkey FOREIGN KEY (team_id1) REFERENCES public.teams(id);


--
-- Name: match_statistics match_statistics_team_id2_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.match_statistics
    ADD CONSTRAINT match_statistics_team_id2_fkey FOREIGN KEY (team_id2) REFERENCES public.teams(id);


--
-- Name: matches matches_team1_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_team1_id_fkey FOREIGN KEY (team1_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: matches matches_team2_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_team2_id_fkey FOREIGN KEY (team2_id) REFERENCES public.teams(id) ON DELETE CASCADE;


--
-- Name: players players_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

