#!/usr/bin/perl

use strict;
use warnings;

use utf8;
use Encode qw(encode_utf8);

use Modern::Perl;

use File::Slurp; 
use Getopt::Compact;

my $TRY_AGAIN_SECONDS = 3;

sub exec_cmd {
	my ($opts, $cmd) = @_;
	
	if ($opts->{sudo}) {
		$cmd = sprintf('sudo "%s"', $cmd);
	}
	
	return `$cmd`;
}

sub options_validate {
	my ($opts) = @_;
	
	if ($opts->{'cache-for'} or $opts->{'cache-to'}) {
		if (not $opts->{'cache-for'} or not $opts->{'cache-to'}) {
			say_error('Arguments cache-for and cache-to are both needed to cache the zypper output');
		
			return;
		}
		
		if ($opts->{'cache-for'} !~ m/^\d+$/) {
			say_error('cache-for must be a positive integer');

			return;
		}
		
# 		if (not -w $opts->{'cache-to'}) {
# 			say_error('cache-to file ' . $opts->{'cache-to'} . ' is not writeable');
# 			
# 			return;
# 		}
	}

	return 1;
}

sub say_error {
	my ($text) = @_;

	say "\x1B[0;31mERROR: $text\x1b[0m\n";
}

my $options = Getopt::Compact->new(
	name => 'Zypper check for check_mk agent',
	struct => [
		[ 'cache-for', 'Cache the output for x seconds', ':i' ],
		[ 'cache-to', 'Cache the output to this file', ':s' ],
		[ 'sudo', 'Sudo every command' ],
		[ 'verbose', 'Verbose output' ],
	]
);

my $opts = $options->opts();

if (not $options->status() or not options_validate($opts)) {
	say $options->usage();

	exit 1;
}

my $zypper_out;

if ($opts->{'cache-to'} and -f $opts->{'cache-to'} and time - (stat $opts->{'cache-to'})[9] < $opts->{'cache-for'}) {
	say 'Read in cached zypper updates' if $opts->{verbose};

	$zypper_out = read_file($opts->{'cache-to'});
}
else {
	while (1) {
		say 'Read new zypper updates' if $opts->{verbose};

		$zypper_out = exec_cmd($opts, 'zypper --non-interactive --no-gpg-checks lu 2>&1');
		
		if ($opts->{'cache-to'}) {
			write_file($opts->{'cache-to'}, $zypper_out);
		}
		
		if ($zypper_out =~ m/System management is locked/sg) {
			say "System management is locked. Will try again in $TRY_AGAIN_SECONDS seconds." if $opts->{verbose};

			sleep $TRY_AGAIN_SECONDS;
			
			next;
		}
		
		last;
	}
}

my $updates;


if ($zypper_out =~ m/No updates found/sg) {
	$updates = 0;
}
else {
	my @lines = split(/\n/, $zypper_out);

	while (my $line = shift @lines) {
		if ($line =~ m/^S\s+\|\s+Repository/) {
			shift @lines; # remove dividing line
			
			last;
		}
	}
	
	$updates = scalar @lines;
}

print "<<<zypper>>>\n";
print "Updates:$updates\n";

exit 0;
