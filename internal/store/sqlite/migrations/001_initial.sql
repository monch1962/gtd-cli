-- Migration: 001_initial.sql
-- Initial schema for GTD CLI

CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    note TEXT,
    status TEXT NOT NULL DEFAULT 'inbox',
    project_id TEXT REFERENCES projects(id),
    area_id TEXT REFERENCES areas(id),
    waiting_for TEXT,
    due_at DATETIME,
    start_at DATETIME,
    tickle_at DATETIME,
    completed_at DATETIME,
    source TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    note TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    area_id TEXT REFERENCES areas(id),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE contexts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE areas (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE task_contexts (
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    context_id TEXT NOT NULL REFERENCES contexts(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, context_id)
);

CREATE TABLE reviews (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    ended_at DATETIME,
    note TEXT,
    stats TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- Indexes for common queries
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_due_at ON tasks(due_at);
CREATE INDEX idx_task_contexts_context_id ON task_contexts(context_id);
