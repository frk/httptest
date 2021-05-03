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

	////////////////////////////////////////////////////////////////////////////
	// Sidebar
	////////////////////////////////////////////////////////////////////////////

	function Sidebar() {

		function itemOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let item = e.currentTarget;
			this.selectItem(item);
			this.expandSubItems(item);
			this.main.scrollIntoView(item);
		}

		return {
			// A reference to the an instance of Main.
			main: null,
			// A map of sidebar list items. The key in the map is the href attribute
			// of the list item's anchor child element.
			itemMap: new Map(),
			// An ordered list of sidebar list items currently shown/expanded.
			shownItems: [],
			// The currently active sidebar list item.
			activeItem: null,

			// init initializes the state of the sidebar.
			init: function() {

				let items = document.getElementsByClassName('sidebar-list-item');
				for (let i = 0; i < items.length; i++) {
					let item = items[i];

					// add an event listener to each sidebar item
					item.addEventListener('click', itemOnClickHandler.bind(this));

					// add to the map
					let a = firstElementChild(item);
					this.itemMap.set(a.getAttribute('href'), item);

					// The server may set an initial item to active,
					// if it did then retain that item for later use.
					if (item.classList.contains('active')) {
						this.activeItem = item;
					}

					// If the item is expanded and has children,
					// aggregate them into the shownItems array.
					if (item.classList.contains('has-subitems')) {
						let subitems = lastElementChild(item);
						if (!subitems.classList.contains('hidden')) {
							this.shownItems.push(subitems);
						}
					}
				}
			},

			// selectItem deactivates any previously selected item
			// and activates the given item.
			selectItem: function(item) {
				if (item !== this.activeItem) {
					if (this.activeItem) {
						this.activeItem.classList.remove('active');
					}

					this.activeItem = item;
					this.activeItem.classList.add('active');
				}
			},

			// expandSubItems displays the sub-items of the given item and
			// hides previously displayed items that aren't in the hierarchy
			// of the given item.
			expandSubItems: function(item) {
				if (!item.classList.contains('has-subitems')) {
					// If item is child of the currently shown item's last element; exit
					if (this.shownItems.length > 0) {
						let last = this.shownItems[this.shownItems.length -1];
						for (let i = 0; i < last.children.length; i++) {
							if (last.children[i] === item) {
								return;
							}
						}
					}
					this.hideShownSiblingChildren(item);
					return;
				}

				// Get the ul element to be displayed.
				let subitems = lastElementChild(item);
				if (subitems === null) {
					return;
				}

				// If member of the current shown items; don't do anything
				if (!subitems.classList.contains('hidden')) {
					return;
				}

				// If currently no subitems are being shown; show subitems
				// and add them to the shownItems array.
				if (this.shownItems.length === 0) {
					subitems.classList.remove('hidden');
					this.shownItems.push(subitems);
					return;
				}

				// If child of the currently shown items' last element; show
				// current subitems and add them to the shownItems array.
				let last = this.shownItems[this.shownItems.length -1];
				for (let i = 0; i < last.children.length; i++) {
					if (last.children[i] === item) {
						subitems.classList.remove('hidden');
						this.shownItems.push(subitems);
						return;
					}
				}

				this.hideShownSiblingChildren(item);
				subitems.classList.remove('hidden');
				this.shownItems.push(subitems);
				return;
			},

			// hideShownSiblingChildren ...
			hideShownSiblingChildren: function(target) {
				let num = this.shownItems.length;
				for (let i = num - 1; i >= 0; i--) {
					let last = this.shownItems.pop()
					last.classList.add('hidden');

					let li = last.parentElement;
					if (li !== null && li.parentElement === target.parentElement) {
						break;
					}
				}
			},
		};
	}

	////////////////////////////////////////////////////////////////////////////
	// Main
	////////////////////////////////////////////////////////////////////////////

	function Main() {

		// endpointItemOnClickHandler
		function endpointItemOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let a = firstElementChild(e.currentTarget);
			let item = this.sidebar.itemMap.get(a.getAttribute('href'));
			this.sidebar.selectItem(item);
			this.sidebar.expandSubItems(item);
			this.scrollIntoView(item);
		}
		
		// fieldListHeadingOnClickHandler
		function fieldListHeadingOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let container = e.currentTarget.parentElement;
			if (container) {
				if (container.classList.contains('collapsed')) {
					container.classList.remove('collapsed');
				} else {
					container.classList.add('collapsed');
				}
			}
		}

		return {
			// A reference to the Sidebar instance.
			sidebar: null,

			// init initializes the state of the main content.
			init: function() {
				this.expandAnchoredField();

				// add an event listener to each endpoint link
				let endpointItems = document.getElementsByClassName('xs-endpoint-item');
				for (let i = 0; i < endpointItems.length; i++) {
					endpointItems[i].addEventListener('click', endpointItemOnClickHandler.bind(this));
				}
				// add an event listener to each child field list
				let fieldLists = document.getElementsByClassName('field-list-container child');
				for (let i = 0; i < fieldLists.length; i++) {
					let heading = firstElementChild(fieldLists[i]);
					heading.addEventListener('click', fieldListHeadingOnClickHandler.bind(this));
				}
			},

			// scrollIntoView ...
			scrollIntoView: function(item) {
				let a = document.getElementById(item.dataset.anchor);
				if (a !== null) {
					a.scrollIntoView();

					let c = firstElementChild(item);
					window.history.pushState({}, "", c.getAttribute('href'));
				}
			},

			// expandAnchoredField ...
			expandAnchoredField: function() {
				if (window.location.hash.length > 0) {
					// slice(1) to drop the leading '#'
					let id = window.location.hash.slice(1);
					let elem = document.getElementById(id);
					if (elem === null) {
						return;
					}

					if (elem.tagName === "LI" && elem.classList.contains('field-item')) {
						for (let el = elem; el !== null; ) {
							// li.field-item -> ul.field-list -> div.field-list-container
							if (el.parentElement && el.parentElement.parentElement) {
								el = el.parentElement.parentElement;

								let cl = el.classList;
								if (cl && cl.contains('field-list-container') && cl.contains('collapsed')) {
									cl.remove('collapsed');

									// div.field-list-container -> li.field-item
									el = el.parentElement;
									continue;
								}
							}

							el = null; // done
						}
					}

					elem.scrollIntoView();
				}
			},
		};
	}

	/////////////////////////////////////////////////////////////////////////////
	// CodeSnippets
	/////////////////////////////////////////////////////////////////////////////
	function CodeSnippets() {

		// langSelectOnChangeHandler
		function langSelectOnChangeHandler(e) {
			let lang = e.currentTarget.value;
			if (this.lang === lang) {
				return;
			}
			
			for (let i = 0; i < this.items.length; i++) {
				let item = this.items[i];
				item.lang[this.lang].classList.remove('selected');
				item.lang[lang].classList.add('selected');

				if (item.sel !== e.currentTarget) {
					item.sel.value = lang;
				}
			}

			this.lang = lang;
			window.history.pushState({}, "", '?lang='+this.lang);
		}

		return {
			// An array of objects that hold a reference to every language
			// select element alongside a plain object that maps the languages
			// to their corresponding code snippet elements.
			// {
			//		sel: <select_element>,
			//		lang: {
			//			"<language>": <code_snippet_element>,
			//			...
			//		},
			// }
			items: [],
			// Holds reference to the currently selected language.
			lang: null,

			// init initializes the state of the code snippets
			init: function() {

				// add an event listener to each lang select
				let langSelects = document.getElementsByClassName('xs-request-lang-select-container');
				for (let i = 0; i < langSelects.length; i++) {
					let s = firstElementChild(langSelects[i]);
					s.addEventListener('change', langSelectOnChangeHandler.bind(this));

					let item = {sel: s, lang: {}};

					let container = langSelects[i].parentElement.parentElement; // xs-request-topbar -> xs-request-container
					if (container !== null) {
						// aggregate code-snippets maps into an array
						let items = container.getElementsByClassName('code-snippet-container');
						for (let i = 0; i < items.length; i++) {
							item.lang[items[i].dataset.lang] = items[i];
						}
					}

					this.items.push(item);
				}
			},

		};
	}

	function HttpDoc() {
		this.main = new Main();
		this.sidebar = new Sidebar();
		this.codeSnippets = new CodeSnippets();

		this.main.sidebar = this.sidebar;
		this.sidebar.main = this.main;

		function init(opts) {
			opts = opts || {};
			this.codeSnippets.lang = opts.lang || 'http';

			this.main.init();
			this.sidebar.init();
			this.codeSnippets.init();
		}

		return { init: init.bind(this) };
	}

	return new HttpDoc();
}());
