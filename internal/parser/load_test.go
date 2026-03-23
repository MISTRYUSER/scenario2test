package parser

import "testing"

func TestLoadScenarioBytes(t *testing.T) {
	input := []byte(`scenario:
  name: checkout flow
  steps:
    - action: open_page
      target: /checkout
    - action: click
      target: pay_now
`)

	scenario, err := LoadScenarioBytes(input)
	if err != nil {
		t.Fatalf("LoadScenarioBytes returned error: %v", err)
	}

	if scenario.Name != "checkout flow" {
		t.Fatalf("expected scenario name %q, got %q", "checkout flow", scenario.Name)
	}

	if len(scenario.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(scenario.Steps))
	}

	if scenario.Steps[0].ID == "" || scenario.Steps[1].ID == "" {
		t.Fatalf("expected normalized step IDs, got %+v", scenario.Steps)
	}
}

func TestLoadScenarioBytesSupportsFlexibleDSL(t *testing.T) {
	input := []byte(`scenario:
  name: E-Commerce Checkout Path Discovery
  steps:
    - action: open_url
      name: 访问首页
      params:
        url: https://mall.com
      next: check_auth
    - action: conditional_branch
      name: 检查登录状态
      branches:
        - condition: auth == 'unlogged'
          next: user_login
        - condition: auth == 'logged'
          next: search_item
    - action: user_login
      name: 登录流程
      params:
        user: demo
        pwd: secret
      next: search_item
    - action: search_item
      name: 搜索商品
      params:
        keyword: MacBook
      next: add_to_cart
    - action: add_to_cart
      name: 加入购物车
      next: terminal_confirm
    - action: terminal_confirm
      name: 结算并断言
      type: end
`)

	scenario, err := LoadScenarioBytes(input)
	if err != nil {
		t.Fatalf("LoadScenarioBytes returned error: %v", err)
	}

	if scenario.Steps[0].Action != "open_page" {
		t.Fatalf("expected open_url to normalize to open_page, got %q", scenario.Steps[0].Action)
	}

	if scenario.Steps[0].Target != "https://mall.com" {
		t.Fatalf("expected params.url to normalize into target, got %q", scenario.Steps[0].Target)
	}

	if len(scenario.Steps[1].Branches) != 2 {
		t.Fatalf("expected step branches to be preserved, got %d", len(scenario.Steps[1].Branches))
	}

	if scenario.Steps[5].Type != "end" {
		t.Fatalf("expected terminal node type to be preserved, got %q", scenario.Steps[5].Type)
	}
}
