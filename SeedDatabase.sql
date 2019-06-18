-- Clear the tables
DELETE FROM slackstockbot.alertsubscription WHERE id > 0;
DELETE FROM slackstockbot.stockprice WHERE symbol IS NOT NULL;

INSERT INTO slackstockbot.stockprice(symbol, price, time) VALUES
('MSFT', 133.01, current_date),
('INTC', 45.32, current_date),
('GE', 9.45, current_date),
('UBER', 52.19, current_date),
('IBM', 135.97, current_date);

INSERT INTO slackstockbot.alertsubscription(slackuser, channel, symbol, targetprice, direction) VALUES
('UKBM681GV', '#slack-api', 'MSFT', 130.0, 'ABOVE'),
('UKBM681GV', '#slack-api', 'INTC', 45.0, 'BELOW'),
('UKBM681GV', '#slack-api', 'GE', 10.0, 'BELOW'),
('UKBM681GV', '#slack-api','UBER', 50.0, 'ABOVE'),
('UKBM681GV', '', 'IBM', 150.0, 'BELOW');

SELECT a.slackuser, a.channel, a.symbol, a.targetprice, a.direction, a.wasnotified, p.price
	FROM slackstockbot.alertsubscription a, slackstockbot.stockprice p
	WHERE a.wasnotified = false AND a.symbol = p.symbol AND
		( (a.direction = 'ABOVE' AND p.price >= a.targetprice) OR (a.direction = 'BELOW' AND p.price <= a.targetprice) )

