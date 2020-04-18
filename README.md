# munn
[![Build Status](https://travis-ci.com/Shamus03/munn.svg?branch=master)](https://travis-ci.com/Shamus03/munn)

CLI tool to project financial portfolio value.

Output is formatted with tabs to easily paste into Excel:

```bash
λ munn example.munn | tail
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
λ munn --image example.munn
```
![](cmd/munn/example.png)


Use the `--debug` flag to debug account changes:
```bash
λ munn --debug --image example.munn | tail
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

Supply a retirement plan with the `--retire` flag to see a projected retirement date:
```bash
λ munn example.munn --years 100 --retire 2080-01-01:25000 | tail
2119-11-30      Retirement      124900.00
2119-12-02      Bank    322148.31
2119-12-02      Savings 242800.00
2119-12-02      Investment      4000.00
2119-12-02      Retirement      124900.00
2119-12-07      Bank    322698.31
2119-12-07      Savings 242800.00
2119-12-07      Investment      4000.00
2119-12-07      Retirement      124900.00
Retirement date: 2068-01-02
```