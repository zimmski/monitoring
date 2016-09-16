#!/usr/bin/perl

use Modern::Perl;

use Getopt::Compact;

my $out = ``;

my $options = Getopt::Compact->new(
	name => 'Check with traceroute',
	struct => [
		[ 'host', 'Host to trace to', '=s' ],
		[ 'find', 'Optional comma separated list of needed hosts in the trace', ':s' ],
		[ 'max-hops', 'Max hops of the trace', ':i' ],
	]
);

my $opts = $options->opts();

if (not $options->status()) {
	say $options->usage();

	exit 1;
}

if (not $opts->{host}) {
	say 'ERROR: host not defined';

	exit 2;
}

my %find;
my $host = $opts->{host};
my $max_hops = $opts->{'max-hops'};

$max_hops ||= 4;

if ($opts->{find}) {
	for my $i(split(/,/, $opts->{find})) {
		$find{$i} = 1;
	}
}
else {
	$find{$host} = 1;
}

my $traceroute_cmd = sprintf('/usr/sbin/traceroute -m %d %s', $max_hops, $host);
my $traceroute_out = `$traceroute_cmd`;

for my $i(keys %find) {
	if ($traceroute_out !~ m/$i/sg) {
		say "CRITICAL | Host '$i' cannot be found";

		exit(2);
	}
}

say 'OK | Tracepath is ok';
exit(0)
