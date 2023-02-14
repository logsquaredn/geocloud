select step_id, job_id, task_type, job_args
from step
where job_id = $1;
