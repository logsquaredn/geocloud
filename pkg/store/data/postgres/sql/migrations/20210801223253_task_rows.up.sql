BEGIN;

INSERT INTO task (
    task_type,
    task_kind,
    task_params
) VALUES (
    'buffer',
    'transformation',
    ARRAY['buffer-distance', 'quadrant-segment-count']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_kind,
    task_params
) VALUES (
    'filter',
    'transformation',
    ARRAY['filter-column', 'filter-value']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_kind,
    task_params
) VALUES (
    'reproject',
    'transformation',
    ARRAY['target-projection']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_kind
) VALUES (
    'removebadgeometry',
    'transformation'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_kind,
    task_params
) VALUES (
    'vectorlookup', 
    'lookup',
    ARRAY['attributes', 'longitude', 'latitude']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_kind,
    task_params
) VALUES (
    'rasterlookup',
    'lookup',
    ARRAY['bands', 'longitude', 'latitude']
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_kind,
    task_params
) VALUES (
    'polygonVectorLookup', 
    'lookup',
    ARRAY['attributes', 'polygon']
) ON CONFLICT DO NOTHING;

COMMIT;
