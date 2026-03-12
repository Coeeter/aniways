CREATE TYPE desktop_platform AS ENUM (
    'darwin-arm64',
    'darwin-x64',
    'win32-x64',
    'win32-arm64',
    'linux-x64',
    'linux-arm64'
);

CREATE TABLE desktop_releases (
    id varchar(21) PRIMARY KEY DEFAULT generate_nanoid (),
    version varchar(20) NOT NULL,
    platform desktop_platform NOT NULL,
    download_url text NOT NULL,
    file_name varchar(255) NOT NULL,
    file_size bigint NOT NULL,
    release_notes text,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (version, platform)
);

CREATE INDEX idx_desktop_releases_version ON desktop_releases (version DESC);

