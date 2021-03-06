#!/usr/bin/perl

use Modern::Perl;

use Mojo::UserAgent;
use Mojo::UserAgent::CookieJar;

my $ua = Mojo::UserAgent->new;

$ua->connect_timeout(45);

$ua = $ua->cookie_jar(Mojo::UserAgent::CookieJar->new);

my $tx = $ua->post($ARGV[0] . '/j_acegi_security_check' => form => { j_username => $ARGV[1], j_password => $ARGV[2] });

if (my $res = $tx->success) {
	if ($tx->res->code == 302 and $tx->res->headers->header('Location') !~ m/error/i) {
		print("OK | Login into Jenkins was successful\n");
		exit(0);
	} else {
		print("WARNING | Login into Jenkins failed, wrong login credentials\n");
		exit(1);
	}
}

print("CRITICAL | Login into Jenkins failed: " . $tx->error . "\n");
exit(2);
