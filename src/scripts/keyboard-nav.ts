/**
 * Keyboard Navigation Module - Functional keyboard navigation
 * Uses pure functions without classes or OOP patterns
 */

// Types
export interface NavState {
  focusedIndex: number;
  items: HTMLElement[];
}

// Pure function: Check if element is a typing context
export const isTypingContext = (target: EventTarget | null): boolean => {
  if (!target || !(target instanceof HTMLElement)) return false;
  
  const tagName = target.tagName.toLowerCase();
  if (tagName === 'input' || tagName === 'textarea') return true;
  if (target.isContentEditable) return true;
  if (target.closest('.search-bar')) return true;
  
  return false;
};

// Pure function: Calculate next index (bounded)
export const getNextIndex = (currentIndex: number, maxIndex: number): number => {
  return Math.min(currentIndex + 1, maxIndex);
};

// Pure function: Calculate previous index (bounded)
export const getPrevIndex = (currentIndex: number): number => {
  return Math.max(currentIndex - 1, 0);
};

// Pure function: Get announcement text for screen reader
export const getAnnouncementText = (item: HTMLElement | undefined): string => {
  if (!item) return '';
  const title = item.querySelector('h2, h3')?.textContent?.trim() || 'Item';
  return `Focused: ${title}`;
};

// Side effect: Remove focus class from all items
const clearFocusClasses = (items: HTMLElement[]): void => {
  items.forEach(item => item.classList.remove('kbd-focused'));
};

// Side effect: Add focus class to item
const addFocusClass = (item: HTMLElement): void => {
  item.classList.add('kbd-focused');
};

// Side effect: Scroll to item
const scrollToItem = (item: HTMLElement): void => {
  const headerHeight = document.querySelector('.site-header')?.clientHeight || 0;
  
  // Get absolute position in document
  const itemRect = item.getBoundingClientRect();
  const absoluteTop = itemRect.top + window.scrollY;
  
  // Scroll so item appears at top of viewport (below header) with small breathing room
  const targetScrollY = absoluteTop - headerHeight - 8;
  
  // Use instant scroll for vim-like behavior - smooth scroll causes race conditions
  // when user rapidly presses j/k keys (getBoundingClientRect() samples mid-animation)
  window.scrollTo({
    top: targetScrollY,
    behavior: 'instant',
  });
};

// Side effect: Announce to screen reader
const announceToScreenReader = (text: string): void => {
  const liveRegion = document.getElementById('kbd-live-region');
  if (liveRegion) {
    liveRegion.textContent = text;
  }
};

// Side effect: Focus an item by index
export const focusCard = (index: number, items: HTMLElement[]): void => {
  if (items.length === 0 || index < 0 || index >= items.length) return;
  
  const item = items[index];
  clearFocusClasses(items);
  addFocusClass(item);
  scrollToItem(item);
  announceToScreenReader(getAnnouncementText(item));
};

// Side effect: Open focused item's link
const openFocused = (index: number, items: HTMLElement[]): void => {
  const item = items[index];
  if (!item) return;
  
  const link = item.querySelector('a[href*="flathub.org"], a[href*="formulae.brew.sh"], a[href*="github.com"]');
  if (link) {
    window.open((link as HTMLAnchorElement).href, '_blank', 'noopener,noreferrer');
  }
};

// Side effect: Focus search input
const focusSearch = (searchInput: HTMLElement | null): void => {
  if (searchInput instanceof HTMLInputElement) {
    searchInput.focus();
    searchInput.select();
  }
};

// Side effect: Show help modal
const showHelp = (): void => {
  document.getElementById('keyboard-help-modal')?.classList.add('visible');
  document.getElementById('keyboard-help-backdrop')?.classList.add('visible');
};

// Side effect: Hide help modal
const hideHelp = (): void => {
  document.getElementById('keyboard-help-modal')?.classList.remove('visible');
  document.getElementById('keyboard-help-backdrop')?.classList.remove('visible');
};

// Side effect: Toggle theme
const toggleTheme = (): void => {
  if (typeof (window as any).toggleTheme === 'function') {
    (window as any).toggleTheme();
  }
};

// Side effect: Scroll page
const scrollPage = (amount: number): void => {
  window.scrollBy({ top: amount, behavior: 'smooth' });
};

// Side effect: Scroll to top
const scrollToTop = (): void => {
  window.scrollTo({ top: 0, behavior: 'smooth' });
};

// Handle keyboard events
export const handleKeyPress = (
  event: KeyboardEvent,
  state: NavState,
  searchInput: HTMLElement | null,
  updateState: (newState: Partial<NavState>) => void
): void => {
  // Skip if typing
  if (isTypingContext(event.target)) return;
  
  const { focusedIndex, items } = state;
  
  switch (event.key) {
    case 'j': // Move down
      event.preventDefault();
      if (items.length > 0) {
        const newIndex = getNextIndex(focusedIndex, items.length - 1);
        updateState({ focusedIndex: newIndex });
        focusCard(newIndex, items);
      }
      break;
      
    case 'k': // Move up
      event.preventDefault();
      if (items.length > 0) {
        const newIndex = getPrevIndex(focusedIndex);
        updateState({ focusedIndex: newIndex });
        focusCard(newIndex, items);
      }
      break;
      
    case '/': // Focus search
    case 's':
      event.preventDefault();
      focusSearch(searchInput);
      break;
      
    case 'o': // Open focused item
    case 'Enter':
      event.preventDefault();
      openFocused(focusedIndex, items);
      break;
      
    case '?': // Show help
      event.preventDefault();
      showHelp();
      break;
      
    case 't': // Toggle theme
      event.preventDefault();
      toggleTheme();
      break;
      
    case ' ': // Page up/down
      event.preventDefault();
      if (event.shiftKey) {
        scrollPage(-window.innerHeight);
      } else {
        scrollPage(window.innerHeight);
      }
      break;
      
    case 'h': // Home (scroll to top)
      event.preventDefault();
      scrollToTop();
      break;
      
    case 'Escape': // Clear focus or close modal
      handleEscape(focusedIndex, items, searchInput);
      break;
  }
};

// Handle escape key
const handleEscape = (
  focusedIndex: number,
  items: HTMLElement[],
  searchInput: HTMLElement | null
): void => {
  const helpModal = document.getElementById('keyboard-help-modal');
  
  // If help modal is open, close it
  if (helpModal?.classList.contains('visible')) {
    hideHelp();
    return;
  }
  
  // If search is focused, blur it
  if (document.activeElement === searchInput) {
    (searchInput as HTMLInputElement)?.blur();
    return;
  }
  
  // Clear focus
  clearFocusClasses(items);
};

// Refresh items list
export const refreshItems = (itemSelector: string): HTMLElement[] => {
  const items = Array.from(document.querySelectorAll(itemSelector)) as HTMLElement[];
  console.log(`[KeyboardNav] Refreshed, now tracking ${items.length} items`);
  return items;
};

// Initialize keyboard navigation
export const initKeyboardNav = (
  itemSelector: string,
  searchInputSelector: string
): void => {
  let state: NavState = {
    focusedIndex: -1,
    items: refreshItems(itemSelector),
  };
  
  const searchInput = document.querySelector(searchInputSelector) as HTMLElement | null;
  
  // Update state helper
  const updateState = (newState: Partial<NavState>): void => {
    state = { ...state, ...newState };
  };
  
  // Keyboard event handler
  const handleKeydown = (e: KeyboardEvent): void => {
    handleKeyPress(e, state, searchInput, updateState);
  };
  
  // Attach event listener
  document.addEventListener('keydown', handleKeydown);
  
  // Expose refresh function
  (window as any).keyboardNavRefresh = () => {
    const items = refreshItems(itemSelector);
    updateState({ items, focusedIndex: -1 });
  };
  
  console.log(`[KeyboardNav] Initialized with ${state.items.length} items`);
};
