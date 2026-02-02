axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded; charset=UTF-8';
axios.defaults.headers.common['X-Requested-With'] = 'XMLHttpRequest';

axios.interceptors.request.use(
    (config) => {
        // If data is FormData, let browser set multipart boundary
        if (config.data instanceof FormData) {
            config.headers['Content-Type'] = 'multipart/form-data';
            return config;
        }

        // Robust Content-Type detection: scan header objects (possible nested like common/post)
        function headersContainJson(h) {
            if (!h) return false;
            // if it's an object, check its keys
            if (typeof h === 'object') {
                for (const k of Object.keys(h)) {
                    try {
                        const v = h[k];
                        if (typeof v === 'string' && v.toLowerCase().indexOf('application/json') !== -1) return true;
                        if (typeof v === 'object' && headersContainJson(v)) return true;
                    } catch (e) {
                        continue;
                    }
                }
                return false;
            }
            if (typeof h === 'string') {
                return h.toLowerCase().indexOf('application/json') !== -1;
            }
            return false;
        }

        if (headersContainJson(config.headers) || headersContainJson(axios.defaults.headers)) {
            return config;
        }

        config.data = Qs.stringify(config.data, {
            arrayFormat: 'repeat',
        });
        return config;
    },
    (error) => Promise.reject(error),
);

axios.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response) {
            const statusCode = error.response.status;
            // Check the status code
            if (statusCode === 401) { // Unauthorized
                return window.location.reload();
            }
        }
        return Promise.reject(error);
    }
);
