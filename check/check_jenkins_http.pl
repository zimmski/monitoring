#!/usr/bin/perl

use Modern::Perl;

use Mojo::UserAgent;
use Mojo::UserAgent::CookieJar;

my $ua = Mojo::UserAgent->new;

$ua->connect_timeout(45);

$ua = $ua->cookie_jar(Mojo::UserAgent::CookieJar->new);
$ua->max_redirects(3);

my $tx = $ua->post($ARGV[0] . '/j_acegi_security_check' => form => { j_username => $ARGV[1], j_password => $ARGV[2], Submit => 'login' });

if (my $res = $tx->success) {
	my $body = $res->body;
	
	if ($body =~ m/Welcome to Jenkins/ or $body =~ m/Willkommen bei Jenkins/) {
		print("OK | Login into jenkins successfully\n");
		exit(0);
	}
}

print("CRITICAL | Cannot login into jenkins: " . $tx->error . "\n");
exit(2);
