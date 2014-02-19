#!/usr/bin/perl

use Modern::Perl;

use Mojo::UserAgent;
use Mojo::UserAgent::CookieJar;

my $ua = Mojo::UserAgent->new;

$ua->connect_timeout(45);

$ua = $ua->cookie_jar(Mojo::UserAgent::CookieJar->new);

my $tx = $ua->post($ARGV[0] . '/sessions/login' => form => { login => $ARGV[1], password => $ARGV[2], commit => 'Log in' });

if (my $res = $tx->success) {
	if ($tx->res->code == 302 and $tx->res->body !~ m/Authentication failed/i) {
		print("OK | Login into SonarQube was successful\n");
		exit(0);
	} else {
		print("WARNING | Login into SonarQube failed, wrong login credentials\n");
		exit(1);
	}
}

print("CRITICAL | Login into SonarQube failed: " . $tx->error . "\n");
exit(2);
