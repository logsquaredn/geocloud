BEGIN;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'buffer',
    ARRAY['distance', 'segment-count']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'filter',
    ARRAY['filter-column', 'filter-value']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'reproject', 
    ARRAY['target-projection']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type
) VALUES (
    'removebadgeometry'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'vectorlookup', 
    ARRAY['longitude', 'latitude']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'rasterlookup', 
    ARRAY['bands', 'longitude', 'latitude']
) ON CONFLICT DO NOTHING;

COMMIT;
