
var args = require('system').args;
var page = require('webpage').create();

page.clipRect = { top: 0, left: 0, width: 1024, height: 768 };

page.onError = function(msg) {
	console.log('rendering failed: ' + msg);
//	phantom.exit(2);
};

page.viewportSize = {
  width: 1024,
  height: 768
};

page.open(args[1], function(status) {
	if (status === 'success') {
		if (args.length > 2) {
			page.render(args[2]);
		} else {
			console.log(page.renderBase64('PNG'));
		}
		phantom.exit(0);
	} else {
		phantom.exit(1);
	}
});
