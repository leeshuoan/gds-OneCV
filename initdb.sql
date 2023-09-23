DROP DATABASE IF EXISTS school;

CREATE DATABASE school;

\c school

CREATE TABLE teachers (
    teacher_email text PRIMARY KEY
);

CREATE TABLE students (
    student_email text PRIMARY KEY,
    is_suspended boolean DEFAULT false
);

CREATE TABLE registrations (
    registration_id serial PRIMARY KEY,
    teacher_email text REFERENCES teachers(teacher_email),
    student_email text REFERENCES students(student_email),
    UNIQUE (teacher_email, student_email) 
);

CREATE TABLE notifications (
    notification_id serial PRIMARY KEY,
    teacher_email text REFERENCES teachers(teacher_email),
    notification_text text NOT NULL
);

CREATE TABLE mentions (
    mention_id serial PRIMARY KEY,
    notification_id int REFERENCES notifications(notification_id),
    student_id text REFERENCES students(student_email) 
);

INSERT INTO teachers (teacher_email) VALUES
    ('teacherken@gmail.com'),
    ('teacherjoe@gmail.com');

INSERT INTO students (student_email) VALUES
    ('studentjon@gmail.com'),
    ('studenthon@gmail.com'),
    ('commonstudent1@gmail.com'),
    ('commonstudent2@gmail.com'),
    ('student_only_under_teacher_ken@gmail.com'),
    ('studentmary@gmail.com'),
    ('studentbob@gmail.com'),
    ('studentagnes@gmail.com'),
    ('studentmiche@gmail.com');

INSERT INTO registrations (teacher_email, student_email) VALUES
    ('teacherken@gmail.com', 'studentjon@gmail.com'),
    ('teacherken@gmail.com', 'studenthon@gmail.com'),
    ('teacherken@gmail.com', 'commonstudent1@gmail.com'),
    ('teacherjoe@gmail.com', 'studentjon@gmail.com'),
    ('teacherjoe@gmail.com', 'studenthon@gmail.com'),
    ('teacherjoe@gmail.com', 'commonstudent2@gmail.com');
