DROP KEYSPACE  IF EXISTS cgrates;

create keyspace cgrates with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

create user cgrates with password 'CGRateS.org';
grant all on keyspace cgrates to cgrates;
