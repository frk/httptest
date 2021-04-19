(function() {
	console.log('TODO');

	////////////////////////////////////////////////////////////////////////////
	//
	// handle events for sidebar items
	// 
	////////////////////////////////////////////////////////////////////////////
	(function () {
		let items = document.getElementsByClassName('sidebar-item');
		let selected = null;

		for (let i = 0; i < items.length; i++) {
			let item = items[i];

			item.addEventListener('click', function(e) {
				let target = e.target;

				// select
				if (target !== selected) {
					if (selected) {
						selected.classList.remove('selected');
					}

					target.classList.add('selected');
					selected = target;
				}
			});
		}

	}());

}());
