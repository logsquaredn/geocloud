BEGIN;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'buffer',
    ARRAY['distance', 'segmentCount']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'filter',
    ARRAY['filterColumn', 'filterValue']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params
) VALUES (
    'reproject', 
    ARRAY['targetProjection']
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

COMMIT;
