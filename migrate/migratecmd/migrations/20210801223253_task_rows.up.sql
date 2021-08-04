BEGIN;

INSERT INTO task (
    task_type,
    task_params,
    task_queue_name,
    task_ref
) VALUES (
    'buffer',
    ARRAY['distance'],
    'logsquaredn-geocloud-1',
    'logsquaredn/geocloud:task-buffer'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params,
    task_queue_name,
    task_ref
) VALUES (
    'filter',
    ARRAY['filterColumn', 'filterValue'],
    'logsquaredn-geocloud-2',
    'logsquaredn/geocloud:task-filter'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params,
    task_queue_name,
    task_ref
) VALUES (
    'reproject', 
    ARRAY['targetProjection'],
    'logsquaredn-geocloud-1',
    'logsquaredn/geocloud:task-reproject'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_queue_name,
    task_ref
) VALUES (
    'removeBadGeometry',
    'logsquaredn-geocloud-2',
    'logsquaredn/geocloud:task-remove-bad-geometry'
) ON CONFLICT DO NOTHING;

COMMIT;
