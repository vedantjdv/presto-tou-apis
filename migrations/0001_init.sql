-- Create chargers table
CREATE TABLE IF NOT EXISTS chargers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create intervals table
-- start_time and end_time are stored as TIME (HH:MM:SS)
-- days_of_week is a bitmask: 1=Mon, 2=Tue, 4=Wed, 8=Thu, 16=Fri, 32=Sat, 64=Sun (Total 127 for all days)
CREATE TABLE IF NOT EXISTS intervals (
    id SERIAL PRIMARY KEY,
    schedule_id INT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    price_per_kwh DECIMAL(10, 4) NOT NULL,
    days_of_week INT NOT NULL DEFAULT 127,
    created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create charger_schedules table to link chargers to schedules
CREATE TABLE IF NOT EXISTS charger_schedules (
    charger_id INT NOT NULL REFERENCES chargers(id) ON DELETE CASCADE,
    schedule_id INT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    PRIMARY KEY (charger_id, schedule_id)
);
