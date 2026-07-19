package rtlplayground

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ScrapeInformation() (*Information, error) {
	data, err := c.get("/information.json")
	if err != nil {
		return nil, err
	}
	var info Information
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("information.json unmarshal: %w", err)
	}
	return &info, nil
}

func (c *Client) ScrapeStatus() ([]StatusEntry, error) {
	data, err := c.get("/status.json")
	if err != nil {
		return nil, err
	}
	var status []StatusEntry
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("status.json unmarshal: %w", err)
	}
	return status, nil
}

func (c *Client) ScrapeSFPDiag() ([]SFPDiagEntry, error) {
	data, err := c.get("/sfp_diag.json")
	if err != nil {
		return nil, err
	}
	var diag []SFPDiagEntry
	if err := json.Unmarshal(data, &diag); err != nil {
		return nil, fmt.Errorf("sfp_diag.json unmarshal: %w", err)
	}
	return diag, nil
}

func (c *Client) ScrapeSFPEEPROM(slot int) (*SFPEEPROM, error) {
	data, err := c.get(fmt.Sprintf("/sfp_eeprom.json?slot=%d", slot))
	if err != nil {
		return nil, err
	}
	var eeprom SFPEEPROM
	if err := json.Unmarshal(data, &eeprom); err != nil {
		return nil, fmt.Errorf("sfp_eeprom.json unmarshal: %w", err)
	}
	return &eeprom, nil
}

func (c *Client) ScrapeEEE() ([]EEEEntry, error) {
	data, err := c.get("/eee.json")
	if err != nil {
		return nil, err
	}
	var eee []EEEEntry
	if err := json.Unmarshal(data, &eee); err != nil {
		return nil, fmt.Errorf("eee.json unmarshal: %w", err)
	}
	return eee, nil
}

func (c *Client) ScrapeBandwidth() ([]BandwidthEntry, error) {
	data, err := c.get("/bandwidth.json")
	if err != nil {
		return nil, err
	}
	var bw []BandwidthEntry
	if err := json.Unmarshal(data, &bw); err != nil {
		return nil, fmt.Errorf("bandwidth.json unmarshal: %w", err)
	}
	return bw, nil
}

func (c *Client) ScrapeMirror() (*MirrorConfig, error) {
	data, err := c.get("/mirror.json")
	if err != nil {
		return nil, err
	}
	var mirror MirrorConfig
	if err := json.Unmarshal(data, &mirror); err != nil {
		return nil, fmt.Errorf("mirror.json unmarshal: %w", err)
	}
	return &mirror, nil
}

func (c *Client) ScrapeLAG() ([]LAGEntry, error) {
	data, err := c.get("/lag.json")
	if err != nil {
		return nil, err
	}
	var lag []LAGEntry
	if err := json.Unmarshal(data, &lag); err != nil {
		return nil, fmt.Errorf("lag.json unmarshal: %w", err)
	}
	return lag, nil
}

func (c *Client) ScrapeMTU() ([]MTUEntry, error) {
	data, err := c.get("/mtu.json")
	if err != nil {
		return nil, err
	}
	var mtu []MTUEntry
	if err := json.Unmarshal(data, &mtu); err != nil {
		return nil, fmt.Errorf("mtu.json unmarshal: %w", err)
	}
	return mtu, nil
}

func (c *Client) ScrapeVLAN(vid int) (*VLANEntry, error) {
	data, err := c.get(fmt.Sprintf("/vlan.json?vid=%d", vid))
	if err != nil {
		return nil, err
	}
	var vlan VLANEntry
	if err := json.Unmarshal(data, &vlan); err != nil {
		return nil, fmt.Errorf("vlan.json unmarshal: %w", err)
	}
	return &vlan, nil
}

func (c *Client) ScrapeVLANList() ([]VLANListItem, error) {
	data, err := c.get("/vlanlist")
	if err != nil {
		return nil, err
	}
	var list []VLANListItem
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("vlanlist unmarshal: %w", err)
	}
	return list, nil
}

func (c *Client) ScrapeMACTable(startIdx int) ([]L2Entry, error) {
	data, err := c.get(fmt.Sprintf("/l2.json?idx=%d", startIdx))
	if err != nil {
		return nil, err
	}
	var entries []L2Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("l2.json unmarshal: %w", err)
	}
	return entries, nil
}

func (c *Client) DeleteL2Entry(idx int) error {
	data, err := c.get(fmt.Sprintf("/l2_del.json?idx=%d", idx))
	if err != nil {
		return err
	}
	var result L2DeleteResult
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("l2_del.json unmarshal: %w", err)
	}
	if result.Result != 1 {
		return fmt.Errorf("l2_del failed: result=%d", result.Result)
	}
	return nil
}

func (c *Client) ScrapeCounters(port int) ([]CounterEntry, error) {
	data, err := c.get(fmt.Sprintf("/counters.json?port=%d", port))
	if err != nil {
		return nil, err
	}
	var counters []CounterEntry
	if err := json.Unmarshal(data, &counters); err != nil {
		return nil, fmt.Errorf("counters.json unmarshal: %w", err)
	}
	return counters, nil
}

func (c *Client) ScrapeAllMACTable() ([]L2Entry, error) {
	var all []L2Entry
	idx := 0
	for {
		entries, err := c.ScrapeMACTable(idx)
		if err != nil {
			return nil, err
		}
		if len(entries) == 0 {
			break
		}
		all = append(all, entries...)
		last := entries[len(entries)-1]
		idx = last.Index + 1
		if len(entries) < 30 {
			break
		}
	}
	return all, nil
}
