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
    ('teacherken@gmail.com', 'commonstudent1@gmail.com'),
    ('teacherken@gmail.com', 'commonstudent2@gmail.com'),
    ('teacherken@gmail.com', 'student_only_under_teacher_ken@gmail.com'),
    ('teacherjoe@gmail.com', 'commonstudent1@gmail.com'),
    ('teacherjoe@gmail.com', 'commonstudent2@gmail.com');
