/**
 * Filters Module - Functional filtering logic for release cards
 * Uses pure functions without classes or OOP patterns
 */

// Types
export interface FilterState {
  verified: boolean;
  unverified: boolean;
  packageType: string;
  category: string;
  appSet: string;
  days: string;
  hideBeta: boolean;
}

export interface FilterStats {
  visible: number;
  total: number;
  flatpak: number;
  homebrew: number;
  os: number;
}

// Pure function: Check if card matches verification filter
export const matchesVerification = (
  card: HTMLElement,
  verifiedOnly: boolean,
  unverifiedOnly: boolean
): boolean => {
  if (!verifiedOnly && !unverifiedOnly) return true;
  
  const isVerified = card.dataset.verified === 'true';
  if (verifiedOnly && !isVerified) return false;
  if (unverifiedOnly && isVerified) return false;
  
  return true;
};

// Pure function: Check if card matches package type filter
export const matchesPackageType = (
  card: HTMLElement,
  packageType: string
): boolean => {
  if (!packageType) return true;
  return (card.dataset.packageType || '') === packageType;
};

// Pure function: Check if card matches category filter
export const matchesCategory = (
  card: HTMLElement,
  category: string
): boolean => {
  if (!category) return true;
  const categories = card.dataset.categories || '';
  return categories.includes(category);
};

// Pure function: Check if card matches app set filter
export const matchesAppSet = (
  card: HTMLElement,
  appSet: string
): boolean => {
  if (!appSet) return true;
  return (card.dataset.appSet || '') === appSet;
};

// Pure function: Check if card matches date filter
export const matchesDateRange = (
  card: HTMLElement,
  days: string,
  now: Date = new Date()
): boolean => {
  if (!days) return true;
  
  const updatedAt = card.dataset.updatedAt;
  if (!updatedAt) return false;
  
  const updatedDate = new Date(updatedAt);
  const daysDiff = (now.getTime() - updatedDate.getTime()) / (1000 * 60 * 60 * 24);
  
  return daysDiff <= parseInt(days);
};

// Pure function: Check if card matches release type filter (beta/alpha)
export const matchesReleaseType = (
  card: HTMLElement,
  hideBeta: boolean
): boolean => {
  if (!hideBeta) return true;
  
  const releaseTitle = card.dataset.releaseTitle || '';
  const lowerTitle = releaseTitle.toLowerCase();
  
  // Check if title contains beta, alpha, rc (release candidate), or dev
  return !(
    lowerTitle.includes('beta') ||
    lowerTitle.includes('alpha') ||
    lowerTitle.includes('.rc') ||
    lowerTitle.includes('-rc') ||
    lowerTitle.includes('dev')
  );
};

// Pure function: Apply all filters to a single card
export const cardMatchesFilters = (
  card: HTMLElement,
  state: FilterState,
  now: Date = new Date()
): boolean => {
  return (
    matchesVerification(card, state.verified, state.unverified) &&
    matchesPackageType(card, state.packageType) &&
    matchesCategory(card, state.category) &&
    matchesAppSet(card, state.appSet) &&
    matchesDateRange(card, state.days, now) &&
    matchesReleaseType(card, state.hideBeta)
  );
};

// Pure function: Calculate filter statistics
export const calculateStats = (
  cards: HTMLElement[],
  visibleCards: HTMLElement[]
): FilterStats => {
  const stats: FilterStats = {
    visible: visibleCards.length,
    total: cards.length,
    flatpak: 0,
    homebrew: 0,
    os: 0,
  };
  
  visibleCards.forEach(card => {
    const packageType = card.dataset.packageType || '';
    if (packageType === 'flatpak') stats.flatpak++;
    else if (packageType === 'homebrew') stats.homebrew++;
    else if (packageType === 'os') stats.os++;
  });
  
  return stats;
};

// Pure function: Format stats for display
export const formatStats = (stats: FilterStats): string => {
  const parts: string[] = [];
  if (stats.flatpak > 0) parts.push(`${stats.flatpak} Flathub`);
  if (stats.homebrew > 0) parts.push(`${stats.homebrew} Homebrew`);
  if (stats.os > 0) parts.push(`${stats.os} OS`);
  
  return parts.length > 0 ? `(${parts.join(', ')})` : '';
};

// Pure function: Get active filter labels
export const getActiveFilters = (
  state: FilterState,
  getSelectText: (selectId: string) => string
): string[] => {
  const filters: string[] = [];
  
  if (state.verified) filters.push('Verified');
  if (state.unverified) filters.push('Unverified');
  if (state.packageType) filters.push(`Type: ${getSelectText('filter-package-type')}`);
  if (state.category) filters.push(`Category: ${getSelectText('filter-category')}`);
  if (state.appSet) filters.push(`App Set: ${getSelectText('filter-app-set')}`);
  if (state.days) filters.push(`Updated: ${getSelectText('filter-date')}`);
  if (state.hideBeta) filters.push('Hide beta/alpha');
  
  return filters;
};

// Side effect: Update DOM with filter results
const updateCardVisibility = (card: HTMLElement, visible: boolean): void => {
  card.style.display = visible ? '' : 'none';
};

// Side effect: Update DOM elements
const updateFilterUI = (
  filterCount: HTMLElement | null,
  filterStats: HTMLElement | null,
  activeFiltersContainer: HTMLElement | null,
  activeFiltersList: HTMLElement | null,
  stats: FilterStats,
  activeFilters: string[]
): void => {
  // Update count
  if (filterCount) {
    filterCount.textContent = `Showing ${stats.visible} of ${stats.total} packages`;
  }
  
  // Update stats breakdown
  if (filterStats) {
    const statsText = formatStats(stats);
    if (statsText) {
      filterStats.textContent = statsText;
      filterStats.style.display = 'block';
    } else {
      filterStats.style.display = 'none';
    }
  }
  
  // Update active filters display
  if (activeFiltersContainer && activeFiltersList) {
    if (activeFilters.length > 0) {
      activeFiltersContainer.style.display = 'block';
      activeFiltersList.innerHTML = activeFilters
        .map(filter => `<span class="active-filter-tag">${filter}</span>`)
        .join('');
    } else {
      activeFiltersContainer.style.display = 'none';
    }
  }
};

// Main filtering function with side effects
export const applyFilters = (
  cards: HTMLElement[],
  state: FilterState,
  elements: {
    filterCount: HTMLElement | null;
    filterStats: HTMLElement | null;
    activeFiltersContainer: HTMLElement | null;
    activeFiltersList: HTMLElement | null;
  },
  getSelectText: (selectId: string) => string
): FilterStats => {
  const now = new Date();
  const visibleCards: HTMLElement[] = [];
  
  // Apply filters to all cards
  cards.forEach(card => {
    const shouldShow = cardMatchesFilters(card, state, now);
    updateCardVisibility(card, shouldShow);
    if (shouldShow) visibleCards.push(card);
  });
  
  // Calculate statistics
  const stats = calculateStats(cards, visibleCards);
  
  // Get active filter labels
  const activeFilters = getActiveFilters(state, getSelectText);
  
  // Update UI
  updateFilterUI(
    elements.filterCount,
    elements.filterStats,
    elements.activeFiltersContainer,
    elements.activeFiltersList,
    stats,
    activeFilters
  );
  
  console.log(
    `[Filters] Applied: ${stats.visible}/${stats.total} visible ` +
    `(${stats.flatpak} Flathub, ${stats.homebrew} Homebrew, ${stats.os} OS)`
  );
  
  return stats;
};

// Initialize filters with event listeners
export const initFilters = (container: HTMLElement): void => {
  // Get DOM elements
  const verifiedCheckbox = document.getElementById('filter-verified') as HTMLInputElement;
  const unverifiedCheckbox = document.getElementById('filter-unverified') as HTMLInputElement;
  const packageTypeSelect = document.getElementById('filter-package-type') as HTMLSelectElement;
  const categorySelect = document.getElementById('filter-category') as HTMLSelectElement;
  const appSetSelect = document.getElementById('filter-app-set') as HTMLSelectElement;
  const dateSelect = document.getElementById('filter-date') as HTMLSelectElement;
  const hideBetaCheckbox = document.getElementById('filter-hide-beta') as HTMLInputElement;
  const clearAllBtn = document.getElementById('clear-all-filters') as HTMLButtonElement;
  
  if (!verifiedCheckbox || !packageTypeSelect || !categorySelect || !appSetSelect || !dateSelect || !hideBetaCheckbox) {
    console.error('[Filters] Required filter elements not found');
    return;
  }
  
  const elements = {
    filterCount: document.getElementById('filter-count'),
    filterStats: document.getElementById('filter-stats'),
    activeFiltersContainer: document.getElementById('active-filters'),
    activeFiltersList: document.getElementById('active-filters-list'),
  };
  
  // Get all cards
  const getAllCards = (): HTMLElement[] => 
    Array.from(document.querySelectorAll('.release-card'));
  
  // Get current filter state
  const getFilterState = (): FilterState => ({
    verified: verifiedCheckbox.checked,
    unverified: unverifiedCheckbox.checked,
    packageType: packageTypeSelect.value,
    category: categorySelect.value,
    appSet: appSetSelect.value,
    days: dateSelect.value,
    hideBeta: hideBetaCheckbox.checked,
  });
  
  // Get selected option text
  const getSelectText = (selectId: string): string => {
    const select = document.getElementById(selectId) as HTMLSelectElement;
    return select?.options[select.selectedIndex]?.text || '';
  };
  
  // Apply filters handler
  const handleApplyFilters = (): void => {
    const cards = getAllCards();
    const state = getFilterState();
    applyFilters(cards, state, elements, getSelectText);
  };
  
  // Verification checkbox handler (mutual exclusivity)
  const handleVerificationChange = (e: Event): void => {
    const target = e.target as HTMLInputElement;
    if (target.checked) {
      if (target.id === 'filter-verified' && unverifiedCheckbox) {
        unverifiedCheckbox.checked = false;
      } else if (target.id === 'filter-unverified' && verifiedCheckbox) {
        verifiedCheckbox.checked = false;
      }
    }
    handleApplyFilters();
  };
  
  // Clear all filters handler
  const handleClearAll = (): void => {
    verifiedCheckbox.checked = false;
    if (unverifiedCheckbox) unverifiedCheckbox.checked = false;
    packageTypeSelect.value = '';
    categorySelect.value = '';
    appSetSelect.value = '';
    dateSelect.value = '';
    hideBetaCheckbox.checked = false;
    handleApplyFilters();
  };
  
  // Attach event listeners
  verifiedCheckbox.addEventListener('change', handleVerificationChange);
  if (unverifiedCheckbox) {
    unverifiedCheckbox.addEventListener('change', handleVerificationChange);
  }
  packageTypeSelect.addEventListener('change', handleApplyFilters);
  categorySelect.addEventListener('change', handleApplyFilters);
  appSetSelect.addEventListener('change', handleApplyFilters);
  dateSelect.addEventListener('change', handleApplyFilters);
  hideBetaCheckbox.addEventListener('change', handleApplyFilters);
  if (clearAllBtn) {
    clearAllBtn.addEventListener('click', handleClearAll);
  }
  
  // Initial filter application
  handleApplyFilters();
  
  // Expose refresh function
  (window as any).refreshFilters = () => {
    handleApplyFilters();
  };
  
  console.log(`[Filters] Initialized with ${getAllCards().length} cards`);
};
