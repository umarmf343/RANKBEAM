const AmazonScraper = require('./Amazon');
const { geo, defaultItemLimit } = require('./constant');

const COUNTRY_ALIASES = {
    UK: 'GB',
};

const COUNTRY_DISPLAY = Object.entries(COUNTRY_ALIASES).reduce((accumulator, [alias, canonical]) => {
    accumulator[canonical] = alias;
    return accumulator;
}, {});

const resolveCountryCode = (country) => {
    if (!country) {
        return 'US';
    }

    const normalized = String(country).trim().toUpperCase();
    const canonical = COUNTRY_ALIASES[normalized] ? COUNTRY_ALIASES[normalized] : normalized;

    if (geo[canonical]) {
        return canonical;
    }

    return 'US';
};

const INIT_OPTIONS = {
    bulk: true,
    number: defaultItemLimit,
    filetype: '',
    rating: [1, 5],
    page: 1,
    category: '',
    cookie: '',
    asyncTasks: 5,
    sponsored: false,
    category: 'aps',
    cli: false,
    sort: false,
    discount: false,
    reviewFilter: {
        // Sort by recent/top reviews
        sortBy: 'recent',
        // Show only reviews with verified purchase
        verifiedPurchaseOnly: false,
        // Show only reviews with specific rating or positive/critical
        filterByStar: '',
        formatType: 'all_formats',
    },
};

exports.products = async (options) => {
    options = { ...INIT_OPTIONS, ...options };
    const countryCode = resolveCountryCode(options.country);
    options.country = countryCode;
    options.geo = geo[countryCode];
    options.scrapeType = 'products';
    if (!options.bulk) {
        options.asyncTasks = 1;
    }
    try {
        const data = await new AmazonScraper(options).startScraper();
        return data;
    } catch (error) {
        throw error;
    }
};

exports.reviews = async (options) => {
    options = { ...INIT_OPTIONS, ...options };
    const countryCode = resolveCountryCode(options.country);
    options.country = countryCode;
    options.geo = geo[countryCode];
    options.scrapeType = 'reviews';
    if (!options.bulk) {
        options.asyncTasks = 1;
    }
    try {
        const data = await new AmazonScraper(options).startScraper();
        return data;
    } catch (error) {
        throw error;
    }
};

exports.asin = async (options) => {
    options = { ...INIT_OPTIONS, ...options };
    const countryCode = resolveCountryCode(options.country);
    options.country = countryCode;
    options.geo = geo[countryCode];
    options.scrapeType = 'asin';
    options.asyncTasks = 1;
    try {
        const data = await new AmazonScraper(options).startScraper();
        return data;
    } catch (error) {
        throw error;
    }
};

exports.categories = async (options) => {
    options = { ...INIT_OPTIONS, ...options };
    const countryCode = resolveCountryCode(options.country);
    options.country = countryCode;
    options.geo = geo[countryCode];
    try {
        const data = await new AmazonScraper(options).extractCategories();
        return data;
    } catch (error) {
        throw error;
    }
};

exports.countries = async () => {
    const output = [];
    for (let item in geo) {
        const displayCode = COUNTRY_DISPLAY[item] ? COUNTRY_DISPLAY[item] : item;
        output.push({
            country: geo[item].country,
            country_code: displayCode,
            currency: geo[item].currency,
            host: geo[item].host,
        });
    }
    return output;
};
