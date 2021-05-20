INSERT INTO
    instance (
        id,
        creator_id,
        updater_id,
        workspace_id,
        environment_id,
        name,
        external_link,
        host,
        port
    )
VALUES
    (
        6001,
        1001,
        1001,
        1,
        5001,
        'On-premise MySQL instance',
        'localhost',
        '127.0.0.1',
        '33060'
    );

INSERT INTO
    instance (
        id,
        creator_id,
        updater_id,
        workspace_id,
        environment_id,
        name,
        external_link,
        host,
        port
    )
VALUES
    (
        6002,
        1001,
        1001,
        1,
        5002,
        'AWS RDS instance',
        'google.com',
        '127.0.0.1',
        ''
    );

INSERT INTO
    instance (
        id,
        creator_id,
        updater_id,
        workspace_id,
        environment_id,
        name,
        external_link,
        host,
        port
    )
VALUES
    (
        6003,
        1001,
        1001,
        1,
        5003,
        'GCP Cloud SQL instance',
        'google.com',
        '13.24.32.122',
        '15202'
    );

INSERT INTO
    instance (
        id,
        creator_id,
        updater_id,
        workspace_id,
        environment_id,
        name,
        external_link,
        host,
        port
    )
VALUES
    (
        6004,
        1001,
        1001,
        1,
        5004,
        'Azure SQL instance',
        'google.com',
        'mydb.com',
        '1234'
    );

INSERT INTO
    instance (
        id,
        creator_id,
        updater_id,
        workspace_id,
        environment_id,
        name,
        external_link,
        host,
        port
    )
VALUES
    (
        6005,
        1001,
        1001,
        1,
        5004,
        'AliCloud RDS instance',
        'google.com',
        'rds.com',
        '5678'
    );