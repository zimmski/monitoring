#!/usr/bin/perl

use strict;
use warnings;

my @curl = ('/usr/bin/curl', '"' . $ARGV[1] . '"', '-s', '--ntlm', '-u', '"' . $ARGV[2] . '"');
my $cmd = join(' ', @curl);
my $out = `$cmd`;

my $found = undef;
my $search = $ARGV[0];

if ($out =~ /$search/){
	$found = 1;
}

if ($found){
	print("OK | Content '$search' has been found\n");
	exit(0);
}
else {
	print("CRITICAL | Content '$search' has not been found\n");
	exit(2);
}

