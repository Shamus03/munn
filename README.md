# munn

CLI tool to project financial portfolio value.

Output is formatted with tabs to easily paste into Excel:

```bash
λ munn example.yaml | tail
2022-11-24      Investment      4000.00
2022-11-24      Retirement      8600.00
2022-12-01      Bank    12735.98
2022-12-01      Savings 10200.00
2022-12-01      Investment      4000.00
2022-12-01      Retirement      8600.00
2022-12-02      Bank    10800.78
2022-12-02      Savings 10200.00
2022-12-02      Investment      4000.00
2022-12-02      Retirement      8600.00
```

You can also generate a graph image:
```bash
λ munn --image example.yaml
```
![](cmd/munn/example.png)


Use the `-debug` flag to debug account changes:
```bash
λ munn -debug -image example.yaml | tail
2022-12-02, Account Savings gained interest
2022-12-02, Account Investment gained interest
2022-12-02, Account Retirement gained interest
2022-12-02, Applied transaction Phone
2022-12-02, Applied transaction Rent
2022-12-02, Applied transaction Spotify
2022-12-02, Applied transaction Credit Card
2022-12-02, Applied transaction Internet
2022-12-02, Applied transaction Electric
2022-12-02, Applied transaction Auto/Renters Insurance
```