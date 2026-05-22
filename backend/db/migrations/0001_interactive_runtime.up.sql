create table if not exists interactive_units (
    id uuid primary key,
    exam_id uuid not null references exams(id) on delete cascade,
    subject_id uuid null references subjects(id) on delete set null,
    title text not null,
    status text not null default 'draft',
    current_published_version_id uuid null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists interactive_unit_versions (
    id uuid primary key,
    interactive_unit_id uuid not null references interactive_units(id) on delete cascade,
    version_no int not null,
    status text not null default 'draft',
    metadata jsonb not null default '{}'::jsonb,
    published_at timestamptz null,
    published_by uuid null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique(interactive_unit_id, version_no)
);

create table if not exists interactive_unit_version_steps (
    id uuid primary key,
    unit_version_id uuid not null references interactive_unit_versions(id) on delete cascade,
    step_no int not null,
    widget_type text not null,
    content jsonb not null default '{}'::jsonb,
    initial_state jsonb not null default '{}'::jsonb,
    allowed_actions jsonb not null default '{}'::jsonb,
    evaluation_config jsonb not null default '{}'::jsonb,
    feedback_map jsonb not null default '{}'::jsonb,
    hint_policy jsonb not null default '{}'::jsonb,
    knowledge_point_ids jsonb not null default '[]'::jsonb,
    knowledge_point_tags jsonb not null default '[]'::jsonb,
    created_at timestamptz not null default now(),
    unique(unit_version_id, step_no)
);

create table if not exists unit_attempts (
    id uuid primary key,
    user_id uuid not null,
    unit_version_id uuid not null references interactive_unit_versions(id) on delete cascade,
    status text not null default 'in_progress',
    created_at timestamptz not null default now(),
    completed_at timestamptz null
);

create table if not exists step_actions (
    id uuid primary key,
    attempt_id uuid not null references unit_attempts(id) on delete cascade,
    step_id uuid not null references interactive_unit_version_steps(id) on delete cascade,
    action_payload jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create table if not exists step_feedback (
    id uuid primary key,
    attempt_id uuid not null references unit_attempts(id) on delete cascade,
    step_id uuid not null references interactive_unit_version_steps(id) on delete cascade,
    is_correct boolean not null,
    allow_continue boolean not null,
    hint text null,
    created_at timestamptz not null default now()
);

create table if not exists concept_cards (
    id uuid primary key,
    user_id uuid not null,
    attempt_id uuid not null references unit_attempts(id) on delete cascade,
    unit_version_id uuid not null references interactive_unit_versions(id) on delete cascade,
    content jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);
