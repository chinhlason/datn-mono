create table if not exists doctors(
    id uuid,
    doctor_code text,
    password text,
    fullname text,
    email text,
    phone text,
    role text,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists rooms(
    id uuid,
    id_doctor uuid,
    name text,
    bed_number int,
    patient_number int,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists beds(
    id uuid,
    id_room uuid,
    name text,
    status text,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists medical_records(
    id uuid,
    id_patient uuid,
    id_doctor uuid,
    id_updater uuid,
    status text,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists notes(
    id uuid,
    id_record uuid,
    id_doctor uuid,
    content text,
    img_url text,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists devices(
    id uuid,
    serial text,
    warraty int,
    status text,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists patients(
    id uuid,
    patient_code text,
    fullname text,
    ccid text,
    address text,
    dob text,
    gender text,
    phone text,
    relative_name text,
    relative_phone text,
    reason text,
    create_at timestamp,
    update_at timestamp,
    primary key ( (id) )
);

create table if not exists usage_bed(
    id uuid,
    id_bed uuid,
    id_record uuid,
    status text,
    create_at timestamp,
    end_at timestamp,
    primary key ( (id) )
);

create table if not exists usage_device(
    id uuid,
    id_device uuid,
    id_record uuid,
    status text,
    create_at timestamp,
    end_at timestamp,
    primary key ( (id) )
);

create table if not exists record_history(
    id uuid,
    id_record uuid,
    id_doctor uuid,
    content text,
    create_at timestamp,
    primary key ( (id) )
);