ALTER TABLE meetings ADD COLUMN mode text NOT NULL DEFAULT 'internal' CHECK (mode IN ('internal','external'));
ALTER TABLE meetings ADD COLUMN turn_plan jsonb NOT NULL DEFAULT '[]';
ALTER TABLE meetings ADD COLUMN closure jsonb;
ALTER TABLE meetings DROP CONSTRAINT meetings_status_check;
ALTER TABLE meetings ADD CONSTRAINT meetings_status_check CHECK (status IN ('queued','blocked','waiting_external','completed'));
ALTER TABLE meetings ADD CONSTRAINT meetings_mode_state_check CHECK (
    (mode='internal' AND status IN ('queued','blocked')) OR
    (mode='external' AND status IN ('waiting_external','completed'))
);
ALTER TABLE jobs DROP CONSTRAINT jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check CHECK (status IN ('queued','blocked','waiting_external','completed'));
ALTER TABLE jobs DROP CONSTRAINT jobs_type_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_type_check CHECK (type IN ('meeting.execute','meeting.external'));
ALTER TABLE jobs ADD CONSTRAINT jobs_type_state_check CHECK (
    (type='meeting.execute' AND status IN ('queued','blocked')) OR
    (type='meeting.external' AND status IN ('waiting_external','completed'))
);
ALTER TABLE meeting_messages ADD COLUMN prompt text NOT NULL DEFAULT '';
ALTER TABLE meeting_messages ADD COLUMN agent_run_id text NOT NULL DEFAULT '';
ALTER TABLE meeting_messages ADD COLUMN context_hash text NOT NULL DEFAULT '';
ALTER TABLE meeting_messages ADD COLUMN history_through integer NOT NULL DEFAULT 0;
ALTER TABLE meeting_messages ADD COLUMN origin text NOT NULL DEFAULT 'legacy_unknown';
CREATE UNIQUE INDEX message_agent_run ON meeting_messages(meeting_id,agent_run_id) WHERE agent_run_id<>'';
