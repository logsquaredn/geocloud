DELETE FROM task WHERE task.task_type IN (
    'buffer',
    'filter',
    'reproject',
    'removeBadGeometry'
) AND task.task_ref IN (
    'logsquaredn/geocloud:task-buffer',
    'logsquaredn/geocloud:task-filter',
    'logsquaredn/geocloud:task-reproject',
    'logsquaredn/geocloud:task-remove-bad-geometry'
);
