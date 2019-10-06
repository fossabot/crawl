/*
Package crawl is a simple web crawler with single domain scope. It can be limited with a timeout.




					controler --
						|
						|
					 crawler --- Manage workers, filter links, relay validated results to controller
					/  |
				   /   |
			  worker  worker --- Visit an url, scrap for links, return Result to crawler










*/
package crawl
