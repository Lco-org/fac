-- Update password hashes for users
-- admin password: Admin@123
-- Ezy password: 123456

UPDATE users 
SET password_hash = '$2a$10$ioeOfefxUCB7JwfWmABE4eFM1hNumY.hvfwC.PW5ZbRAFGof2aw1q' 
WHERE username = 'admin';

UPDATE users 
SET password_hash = '$2a$10$.I9E1IdXTonsh9k1vWx6C.yz6usiAYqk27SxcwISP15mUDaHGMDQG' 
WHERE username = 'Ezy';

-- Verify the update
SELECT username, email, role, is_active FROM users;
