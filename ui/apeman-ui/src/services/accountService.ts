import apiClient from "./api-client";

const ACCOUNT_BASE = "/accounts";

class AccountService {
    getAllAccounts() {
        const controller = new AbortController();

        const request = apiClient.get<string[]>(ACCOUNT_BASE, {
        signal: controller.signal,
        });

        return {
            request,
            cancel: () => {
                controller.abort();
            },
        };
    }

    getAccountServices(account_id: string) {
        const controller = new AbortController();

        const request = apiClient.get<string[]>(
        `${ACCOUNT_BASE}/${account_id}/services`,
        {
            signal: controller.signal,
        }
        );

        return {
            request,
            cancel: () => {
                controller.abort();
            },
        };
    }
}

export default new AccountService();