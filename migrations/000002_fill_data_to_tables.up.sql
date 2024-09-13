INSERT INTO auth_data (login, pwd_hash, user_id) VALUES
('user1', 'hash1', 1),
('user2', 'hash2', 2),
('user3', 'hash3', 3),
('user4', 'hash4', 4),
('user5', 'hash5', 5),
('user6', 'hash6', 6),
('user7', 'hash7', 7),
('user8', 'hash8', 8),
('user9', 'hash9', 9),
('user10', 'hash10', 10),
('user11', 'hash11', 11),
('user12', 'hash12', 12),
('user13', 'hash13', 13),
('user14', 'hash14', 14),
('user15', 'hash15', 15),
('user16', 'hash16', 16),
('user17', 'hash17', 17),
('user18', 'hash18', 18),
('user19', 'hash19', 19),
('user20', 'hash20', 20);

INSERT INTO items (item_id, item_name, item_amount, item_price, item_description) VALUES
(1, 'Laptop', 50, 100000, 'High performance laptop'),
(2, 'Smartphone', 100, 30000, 'Latest model smartphone'),
(3, 'Headphones', 200, 5000, 'Noise-cancelling over-ear headphones'),
(4, 'Monitor', 75, 15000, '24 inch LED monitor'),
(5, 'Keyboard', 150, 2000, 'Mechanical gaming keyboard'),
(6, 'Mouse', 150, 1500, 'Wireless mouse'),
(7, 'Webcam', 80, 4000, '1080p HD webcam'),
(8, 'Router', 60, 8000, 'Dual-band Wi-Fi router'),
(9, 'External Hard Drive', 40, 12000, '1TB external hard drive'),
(10, 'USB Flash Drive', 200, 1000, '32GB USB flash drive'),
(11, 'Tablet', 30, 25000, '10 inch Android tablet'),
(12, 'Smartwatch', 80, 15000, 'Fitness tracking smartwatch'),
(13, 'Printer', 50, 12000, 'Multifunctional printer'),
(14, 'Graphics Card', 20, 50000, 'High-end graphics card'),
(15, 'Gaming Chair', 25, 30000, 'Ergonomic gaming chair'),
(16, 'Microphone', 100, 3000, 'USB condenser microphone'),
(17, 'Speakers', 60, 7000, 'Bluetooth portable speakers'),
(18, 'Power Bank', 150, 2000, '10000mAh power bank'),
(19, 'VR Headset', 20, 40000, 'Virtual reality headset'),
(20, 'Cable Management Kit', 100, 500, 'Organize your cables');

INSERT INTO public.orders ("time", order_id, items_id, delivery_addr, user_id) VALUES
(NOW(), 1, '{1,2}', '123 Main St, CityA', 1),
(NOW(), 2, '{3,4}', '456 Park Ave, CityB', 2),
(NOW(), 3, '{5,6}', '789 Elm St, CityC', 3),
(NOW(), 4, '{7,8}', '101 Maple St, CityD', 4),
(NOW(), 5, '{9,10}', '202 Oak St, CityE', 5),
(NOW(), 6, '{11,12}', '303 Pine St, CityF', 6),
(NOW(), 7, '{13,14}', '404 Cedar St, CityG', 7),
(NOW(), 8, '{15,16}', '505 Birch St, CityH', 8),
(NOW(), 9, '{17,18}', '606 Walnut St, CityI', 9),
(NOW(), 10, '{19,20}', '707 Palm St, CityJ', 10),
(NOW(), 11, '{1,3,5}', '808 Ash St, CityK', 11),
(NOW(), 12, '{2,4,6}', '909 Spruce St, CityL', 12),
(NOW(), 13, '{7,9,11}', '111 Chestnut St, CityM', 13),
(NOW(), 14, '{12,13,14}', '222 Willow St, CityN', 14),
(NOW(), 15, '{15,16,17}', '333 Fir St, CityO', 15),
(NOW(), 16, '{18,19,20}', '444 Cypress St, CityP', 16),
(NOW(), 17, '{1,2,3}', '555 Maple St, CityQ', 17),
(NOW(), 18, '{4,5,6}', '666 Oak St, CityR', 18),
(NOW(), 19, '{7,8,9}', '777 Pine St, CityS', 19),
(NOW(), 20, '{10,11,12}', '888 Birch St, CityT', 20);