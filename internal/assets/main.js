var httpdoc = (function() {
	'use strict';

	const ANCHOR_TOP_MIN = -200;
	const ANCHOR_TOP_MAX = 100;

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
			this.main.scrollIntoView(item);
		}

		return {
			elem: null,
			items: null,
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
				this.elem = document.querySelector('nav.sidebar');
				this.items = this.elem.getElementsByClassName('sidebar-list-item');
				for (let i = 0; i < this.items.length; i++) {
					let item = this.items[i];

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

			// selectItemByIndex
			selectItemByIndex: function(index) {
				let item = this.items[index];
				if (item) {
					this.selectItem(item);
				}
			},

			// selectItem
			selectItem: function(item) {
				if (item === this.activeItem) {
					return;
				}
				this.showItem(item);
				this.setItemAsActive(item);
			},

			// showItem
			showItem: function(item) {
				for (let i = this.shownItems.length-1; i > -1; i--) {
					let last = this.shownItems[i];
					if (last === item.parentElement || last.parentElement === item) {
						break;
					}
					last.classList.add('hidden');
					this.shownItems.pop();
				}

				if (this.shownItems.length === 0) {
					// expand parent ul; repeat until root (i.e. until not ul.sidebar-item-subitems)
					for (let ul = item.parentElement; ; ) {
						if (!ul.classList.contains('sidebar-item-subitems')) {
							break;
						}

						ul.classList.remove('hidden');
						this.shownItems.unshift(ul);

						ul = ul.parentElement.parentElement;
					}
				}

				// expand item's subitems ul if any
				if (item.classList.contains('has-subitems')) {
					let ul = item.querySelector('ul.sidebar-item-subitems');
					ul.classList.remove('hidden');
					this.shownItems.push(ul);
				}

				// if item's is not visible due to scroll, scroll.
				let er = this.elem.getBoundingClientRect();
				let ir = item.getBoundingClientRect();
				if (ir.top < er.top) {
					this.elem.scrollTop += (ir.top - er.top);
				} else if (ir.bottom > er.bottom) {
					this.elem.scrollTop += (ir.bottom - er.bottom);
				}
			},

			// setItemAsActive
			setItemAsActive: function(item) {
				if (this.activeItem) {
					this.activeItem.classList.remove('active');
				}

				this.activeItem = item;
				this.activeItem.classList.add('active');
			},
		};
	}

	////////////////////////////////////////////////////////////////////////////
	// Main
	////////////////////////////////////////////////////////////////////////////
	function Main() {
		// event handler
		function endpointItemOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let a = firstElementChild(e.currentTarget);
			let item = this.sidebar.itemMap.get(a.getAttribute('href'));
			this.sidebar.selectItem(item);
			this.scrollIntoView(item);
		}

		// event handler
		function articleAnchorOnClickHandler(e) {
			e.preventDefault();
			e.stopPropagation();

			let item = this.sidebar.itemMap.get(e.currentTarget.getAttribute('href'));
			this.sidebar.selectItem(item);
			this.scrollIntoView(item);
		}
		
		// event handler
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

		// event handler
		function enumListHeadingOnClickHandler(e) {
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

		// event handler
		function onScrollHandler(e) {
			if (this.onScrollHandlerDisabled) {
				this.onScrollHandlerDisabled = false;
				return;
			}

			let scrollTop = this.container.scrollTop, scrollUp = false;
			if (scrollTop < this.containerScrollTop) {
				scrollUp = true;
			}
			this.containerScrollTop = scrollTop;

			if (scrollUp) {
				if (this.currentIndex < 1) {
					return; // nowhere to go
				} else if (this.currentIndex > 0 && scrollTop < 20) { // close to top?
					this.currentIndex = 0;
					this.sidebar.selectItemByIndex(0);
					return;
				}

				// TODO the hard coded numbers used here should probably be calculated
				// based on the height of the viewport, the switching from current
				// to prev (and probably next) looks weird when the viewport's height
				// is smaller than "normal".

				let a = this.articles[this.currentIndex].getBoundingClientRect();
				if (a.top < (window.innerHeight - 200)) {
					return; // nothing to do
				}

				let prevIndex = this.currentIndex-1;
				let b = this.articles[prevIndex].getBoundingClientRect();
				if (b.top > (window.innerHeight - 200)) {
					for (; prevIndex > 0; prevIndex--) {
						let c = this.articles[prevIndex-1].getBoundingClientRect();
						if (c.top < (window.innerHeight - 200)) {
							break;
						}
					}
				}
				this.currentIndex = prevIndex;
				this.sidebar.selectItemByIndex(this.currentIndex);
			} else {
				if (this.currentIndex > (this.articles.length-2)) {
					return; // nowhere to go
				}

				let a = this.articles[this.currentIndex].getBoundingClientRect();
				if (a.top > ANCHOR_TOP_MIN) {
					return; // nothing to do
				}

				let nextIndex = this.currentIndex+1;
				let b = this.articles[nextIndex].getBoundingClientRect();
				if (b.top > ANCHOR_TOP_MAX) {
					return; // nothing to do, yet
				}

				// handle potential "jumps"
				if (b.top < ANCHOR_TOP_MIN) {
					let len = this.articles.length;
					for (; nextIndex < len-1; nextIndex++) {
						let c = this.articles[nextIndex+1].getBoundingClientRect();
						if (c.top > ANCHOR_TOP_MAX) {
							break;
						}
					}
				}
				this.currentIndex = nextIndex;
				this.sidebar.selectItemByIndex(this.currentIndex);
			}
		}

		return {
			// A reference to the Sidebar instance.
			sidebar: null,
			articles: null,
			currentIndex: 0,
			container: 0,
			containerScrollTop: 0,
			onScrollHandlerDisabled: false,

			// init initializes the state of the main content.
			init: function() {
				this.articles = document.getElementsByTagName('ARTICLE');

				this.container = document.querySelector('body div.content-container');
				this.containerScrollTop = this.container.scrollTop;
				this.container.addEventListener('scroll', onScrollHandler.bind(this));

				this.expandAnchoredField();

				// add an event listener to each endpoint link
				let endpointItems = document.getElementsByClassName('xs-endpoint-item');
				for (let i = 0; i < endpointItems.length; i++) {
					endpointItems[i].addEventListener('click', endpointItemOnClickHandler.bind(this));
				}
				// add an event listener to each article anchor
				let articleAnchors = document.getElementsByClassName('article-anchor');
				for (let i = 0; i < articleAnchors.length; i++) {
					articleAnchors[i].addEventListener('click', articleAnchorOnClickHandler.bind(this));
				}
				// add an event listener to each child field list
				let fieldLists = document.getElementsByClassName('field-list-container child');
				for (let i = 0; i < fieldLists.length; i++) {
					let heading = firstElementChild(fieldLists[i]);
					heading.addEventListener('click', fieldListHeadingOnClickHandler.bind(this));
				}
				// add an event listener to each enum list
				let enumLists = document.getElementsByClassName('enum-list-container');
				for (let i = 0; i < enumLists.length; i++) {
					let heading = firstElementChild(enumLists[i]);
					heading.addEventListener('click', enumListHeadingOnClickHandler.bind(this));
				}
			},

			// scrollIntoView ...
			scrollIntoView: function(item) {
				this.onScrollHandlerDisabled = true;
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
