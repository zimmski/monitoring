#!/opt/bin/perl

use strict;
use warnings;

#
# CPU info
#

my $cpu_temperature = `getsysinfo cputmp`;
$cpu_temperature =~ s/^.*?(\d+) C.+$/$1/sg;

#
# System info
#

my $system_temperature = `getsysinfo systmp`;
$system_temperature =~ s/^.*?(\d+) C.+$/$1/sg;

#
# Fan info
#

my $fan_count = `getsysinfo sysfannum`;
$fan_count =~ s/^.*(\d+).*$/$1/sg;

my %fan_speeds;

for my $i(1..$fan_count) {
	my $fan_speed = `getsysinfo sysfan $i`;
	
	if ($fan_speed =~ s/^.*?(\d+) RPM.*$/$1/sg) {
		$fan_speeds{$i} = $fan_speed;
	}
}

#
# HDD info
#

my $hdd_count = `getsysinfo hdnum`;
$hdd_count =~ s/^.*(\d+).*$/$1/sg;

my %hdd_info;

for my $i(1..$hdd_count) {
	my $hdd_status = `getsysinfo hdstatus $i`;
	$hdd_status =~ s/^.*(\-?\d+).*$/$1/sg;
	
	if ($hdd_status == 0) {
		my $hdd_model = `getsysinfo hdmodel $i`;
		$hdd_model =~ s/[\n\r]//sg;
		my $hdd_smart = `getsysinfo hdsmart $i`;
		$hdd_smart =~ s/[\n\r]//sg;
		my $hdd_temperature = `getsysinfo hdtmp $i`;
		$hdd_temperature =~ s/^.*?(\d+) C.+$/$1/sg;

		$hdd_info{$i} = {
			model => $hdd_model,
			smart => $hdd_smart,
			temperature => $hdd_temperature,
		};
	}
}

#
# OUTPUT
#

print "<<<qnap-hardware>>>\n";

print sprintf("CPU: temperature=%d\n", $cpu_temperature);

print sprintf("System: temperature=%d\n", $system_temperature);

print sprintf("Fan %d: speed=%d\n", $_, $fan_speeds{$_}) for sort keys %fan_speeds;

print sprintf("HDD %d: model=%s;;smart=%s;;temperature=%d\n", $_, $hdd_info{$_}->{model}, $hdd_info{$_}->{smart}, $hdd_info{$_}->{temperature}) for sort keys %hdd_info;
