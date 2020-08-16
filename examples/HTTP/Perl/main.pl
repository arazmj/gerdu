#!/usr/bin/perl
=begin comment
  GoCachePerl Example
  Created by Amir Razmjou on 8/13/20.
  Copyright © 2020 Amir Razmjou. All rights reserved.
=cut

use strict;
use warnings FATAL => 'all';

use LWP::UserAgent;
use HTTP::Request;

my $url = "http://localhost:8080/cache/Hello";
my $ua = LWP::UserAgent->new;

my $putRequest = HTTP::Request->new(PUT => $url);
$putRequest->content('World');
$ua->request($putRequest);

my $getRequest = HTTP::Request->new(GET => $url);
my $response = $ua->request($getRequest);
my $value = $response->content;
print "Hello = ", $value;

