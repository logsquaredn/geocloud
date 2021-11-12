BEGIN;

INSERT INTO task (
    task_type,
    task_params,
    task_ref
) VALUES (
    'buffer',
    ARRAY['distance', 'quadSegCount'],
    'docker.io/logsquaredn/geocloud:task-buffer'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params,
    task_ref
) VALUES (
    'filter',
    ARRAY['filterColumn', 'filterValue'],
    'docker.io/logsquaredn/geocloud:task-filter'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_params,
    task_ref
) VALUES (
    'reproject', 
    ARRAY['targetProjection'],
    'docker.io/logsquaredn/geocloud:task-reproject'
) ON CONFLICT DO NOTHING;

INSERT INTO task (
    task_type,
    task_ref
) VALUES (
    'removebadgeometry',
    'docker.io/logsquaredn/geocloud:task-remove-bad-geometry'
) ON CONFLICT DO NOTHING;

COMMIT;
