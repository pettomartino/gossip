CREATE TRIGGER words_event
AFTER INSERT OR UPDATE OR DELETE ON words
    FOR EACH ROW EXECUTE PROCEDURE save_event();




CREATE OR REPLACE FUNCTION save_event() RETURNS TRIGGER AS $$
    DECLARE
        data json;
        event_id events.id%TYPE;
    BEGIN

        IF (TG_OP = 'DELETE') THEN
            data = row_to_json(OLD);
        ELSE
            data = row_to_json(NEW);
        END IF;

        INSERT INTO events(table_name, action, notification) values (TG_TABLE_NAME, TG_OP, data) RETURNING id INTO event_id;

        PERFORM pg_notify(TG_TABLE_NAME, json_build_object('event_id', event_id::text));

        RETURN NULL;
    END;

$$ LANGUAGE plpgsql;


CREATE TRIGGER events_event
AFTER INSERT OR UPDATE OR DELETE ON events
    FOR EACH ROW EXECUTE PROCEDURE save_event_subscriber();


CREATE OR REPLACE FUNCTION save_event_subscriber() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO event_subscriptions SELECT id, NEW.id as event_id FROM subscribers where table_name = NEW.table_name;
        RETURN NULL;
    END;

$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION notify_event(table_name VARCHAR, event_id integer) RETURNS VOID AS $$
    BEGIN
        PERFORM pg_notify(table_name, TG_ARGV[0]);

        -- Result is ignored since this is an AFTER trigger
        RETURN NULL;
    END;

$$ LANGUAGE plpgsql;