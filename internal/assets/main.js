(function() {
	console.log('TODO');

	function lastElementChild(e) {
		if (e.children.length > 0) {
			return e.children[e.children.length - 1];
		}
		return null;
	}

	function firstElementChild(e) {
		if (e.children.length > 0) {
			return e.children[0];
		}
		return null;
	}

	/**
	* @param {Element} target is the element whose sibling children should be hidden.
	* @param {Array} siblingChildren is a line of descendant elements one of which
	* is expected to be the sibling of the target.
	* */
	function hideSiblingChildren(target, siblingChildren) {
		let num = siblingChildren.length;
		for (let i = num - 1; i >= 0; i--) {
			let last = siblingChildren.pop()
			last.classList.add('hidden');

			let li = last.parentElement;
			if (li !== null && li.parentElement === target.parentElement) {
				break;
			}
		}
	}

	////////////////////////////////////////////////////////////////////////////
	//
	// handle events for sidebar items
	// 
	////////////////////////////////////////////////////////////////////////////
	(function () {
		let items = document.getElementsByClassName('sidebar-list-item');
		let active = null;
		let shown = [];

		for (let i = 0; i < items.length; i++) {
			if (items[i].classList.contains('active')) {
				active = items[i];
			}
			if (items[i].classList.contains('has-subitems')) {
				let subitems = lastElementChild(items[i]);
				if (!subitems.classList.contains('hidden')) {
					shown.push(subitems);
				}
			}

			items[i].addEventListener('click', function(e) {
				e.preventDefault();
				e.stopPropagation();

				let target = e.currentTarget;

				// select
				if (target !== active) {
					if (active) {
						active.classList.remove('active');
					}
					target.classList.add('active');
					active = target;
				}

				// scroll to selected element
				let child = firstElementChild(active);
				if (child !== null) {
					let url = new URL(child.href);
					let a = document.getElementById(url.hash.slice(1));
					if (a !== null) {
						a.scrollIntoView();
						window.history.pushState({}, "", url.pathname);
					}
				}

				if (!target.classList.contains('has-subitems')) {
				  	// if child of the currently shown items' last element; exit
				  	if (shown.length > 0) {
						let last = shown[shown.length -1];
						for (let i = 0; i < last.children.length; i++) {
							if (last.children[i] === target) {
								return;
							}
						}
				  	}
					hideSiblingChildren(target, shown);
				} else {
					// show subitems ...
				  	let subitems = lastElementChild(target);
				  	if (subitems === null) {
				  		return;
				  	}

				  	// if member of the current shown items; don't do anything
				  	if (!subitems.classList.contains('hidden')) {
				  		return;
				  	}

				  	// if currently no subitems are being shown; show subitems
				  	// and add them to the list
				  	if (shown.length === 0) {
				  		subitems.classList.remove('hidden');
				  		shown.push(subitems);
				  		return;
				  	}

				  	// if child of the currently shown items' last element; show
				  	// current subitems and add them to the list
				  	let last = shown[shown.length -1];
				  	for (let i = 0; i < last.children.length; i++) {
						if (last.children[i] === target) {
							subitems.classList.remove('hidden');
							shown.push(subitems);
							return;
						}
				  	}

					hideSiblingChildren(target, shown);
					subitems.classList.remove('hidden');
					shown.push(subitems);
					return;
				}
			});
		}

	}());

}());
