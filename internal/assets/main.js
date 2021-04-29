var httpdoc = (function() {
	'use strict';

	////////////////////////////////////////////////////////////////////////////
	// Generic Helpers
	////////////////////////////////////////////////////////////////////////////

	function lastElementChild(elem) {
		if (elem.children.length > 0) {
			return elem.children[elem.children.length - 1];
		}
		return null;
	}

	function firstElementChild(elem) {
		if (elem.children.length > 0) {
			return elem.children[0];
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

	function HttpDoc() {
		this.items = new Map();
		this.shown = [];
		this.active = null;
		this.codeSnippets = [];

		function selectItem(item) {
			if (item !== this.active) {
				if (this.active) {
					this.active.classList.remove('active');
				}

				this.active = item;
				this.active.classList.add('active');
			}
		}
		this.selectItem = selectItem;

		function scrollIntoView(item) {
			let a = document.getElementById(item.dataset.anchor);
			if (a !== null) {
				a.scrollIntoView();

				let c = firstElementChild(item);
				window.history.pushState({}, "", c.getAttribute('href'));
			}
		}
		this.scrollIntoView = scrollIntoView;

		function expandSubItems(item) {
			if (!item.classList.contains('has-subitems')) {
				// if child of the currently shown item's last element; exit
				if (this.shown.length > 0) {
					let last = this.shown[this.shown.length -1];
					for (let i = 0; i < last.children.length; i++) {
						if (last.children[i] === item) {
							return;
						}
					}
				}
				hideSiblingChildren(item, this.shown);
				return;
			}

			// show subitems ...
			let subitems = lastElementChild(item);
			if (subitems === null) {
				return;
			}

			// if member of the current shown items; don't do anything
			if (!subitems.classList.contains('hidden')) {
				return;
			}

			// if currently no subitems are being shown; show subitems
			// and add them to the list
			if (this.shown.length === 0) {
				subitems.classList.remove('hidden');
				this.shown.push(subitems);
				return;
			}

			// if child of the currently shown items' last element; show
			// current subitems and add them to the list
			let last = this.shown[this.shown.length -1];
			for (let i = 0; i < last.children.length; i++) {
				if (last.children[i] === item) {
					subitems.classList.remove('hidden');
					this.shown.push(subitems);
					return;
				}
			}

			hideSiblingChildren(item, this.shown);
			subitems.classList.remove('hidden');
			this.shown.push(subitems);
			return;
		}
		this.expandSubItems = expandSubItems;

		function sidebarListItemOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let item = e.currentTarget;
			this.selectItem(item);
			this.scrollIntoView(item);
			this.expandSubItems(item);
		}

		function endpointItemOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let a = firstElementChild(e.currentTarget);
			let item = this.items.get(a.getAttribute('href'));
			this.selectItem(item);
			this.scrollIntoView(item);
			this.expandSubItems(item);
		}

		function langSelectOnChangeHandler(e) {
			let lang = e.currentTarget.value;
			if (this.lang === lang) {
				return;
			}
			
			for (let i = 0; i < this.codeSnippets.length; i++) {
				let item = this.codeSnippets[i];
				item.lang[this.lang].classList.remove('selected');
				item.lang[lang].classList.add('selected');

				if (item.sel !== e.currentTarget) {
					item.sel.value = lang;
					//
					// update each option, is this necessary?
					// 
					// for (let i = 0; i < item.sel.children.length; i++) {
					// 	if (item.sel.children[i].value === lang) {
					// 		item.sel.children[i].selected = true;
					// 	} else {
					// 		item.sel.children[i].selected = false;
					// 	}
					// }
				}
			}

			this.lang = lang;
			window.history.pushState({}, "", '?lang='+this.lang);
		}

		function init(opts) {
			opts = opts || {};
			this.lang = opts.lang || 'http';

			// add listeners to sidebar items
			let items = document.getElementsByClassName('sidebar-list-item');
			for (let i = 0; i < items.length; i++) {
				let item = items[i];
				let a = firstElementChild(item);
				item.addEventListener('click', sidebarListItemOnClickHandler.bind(this));

				this.items.set(a.getAttribute('href'), item);
				if (item.classList.contains('active')) {
					this.active = item;
				}
				if (item.classList.contains('has-subitems')) {
				 	let subitems = lastElementChild(item);
				 	if (!subitems.classList.contains('hidden')) {
				 		this.shown.push(subitems);
				 	}
				}
			}

			// add listeners to endpoint links
			let endpointItems = document.getElementsByClassName('xs-endpoint-item');
			for (let i = 0; i < endpointItems.length; i++) {
				let item = endpointItems[i];
				let a = firstElementChild(item);
				item.addEventListener('click', endpointItemOnClickHandler.bind(this));
			}

			// add listeners to lang selects
			let langSelects = document.getElementsByClassName('xs-request-lang-select-container');
			for (let i = 0; i < langSelects.length; i++) {
				let s = firstElementChild(langSelects[i]);
				s.addEventListener('change', langSelectOnChangeHandler.bind(this));

				let item = {sel: s, lang: {}};

				let container = langSelects[i].parentElement.parentElement; // xs-request-topbar -> xs-request-container
				if (container !== null) {
					// aggregate code-snippets maps into an array
					let codeSnippets = container.getElementsByClassName('code-snippet-container');
					for (let i = 0; i < codeSnippets.length; i++) {
						item.lang[codeSnippets[i].dataset.lang] = codeSnippets[i];
					}
				}

				this.codeSnippets.push(item);
			}
		}

		return { init: init.bind(this) };
	}

	return new HttpDoc();
}());
