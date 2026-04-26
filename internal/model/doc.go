// Package model defines shared crawler data types passed between pipeline stages.
//
// Record examples:
//
// Main page record:
//
//	{
//	  "url": "https://example.com/page",
//	  "urlWithoutAnchor": "https://example.com/page",
//	  "breadcrumbSegments": ["Guides"],
//	  "breadcrumbHierarchy": {
//	    "lvl0": "Guides"
//	  },
//	  "contentType": "guide",
//	  "recordType": "lvl1",
//	  "hierarchy": {
//	    "lvl1": "Page Title"
//	  },
//	  "position": 0,
//	  "objectID": "https:--example.com-page"
//	}
//
// Heading record:
//
//	{
//	  "url": "https://example.com/page#section",
//	  "urlWithoutAnchor": "https://example.com/page",
//	  "breadcrumbSegments": ["Guides"],
//	  "breadcrumbHierarchy": {
//	    "lvl0": "Guides"
//	  },
//	  "contentType": "guide",
//	  "recordType": "lvl2",
//	  "hierarchy": {
//	    "lvl1": "Page Title",
//	    "lvl2": "Section"
//	  },
//	  "position": 1,
//	  "objectID": "https:--example.com-page#section"
//	}
//
// Paragraph record:
//
//	{
//	  "url": "https://example.com/page#section",
//	  "urlWithoutAnchor": "https://example.com/page",
//	  "breadcrumbSegments": ["Guides"],
//	  "breadcrumbHierarchy": {
//	    "lvl0": "Guides"
//	  },
//	  "contentType": "guide",
//	  "recordType": "content",
//	  "content": "Paragraph text",
//	  "hierarchy": {
//	    "lvl1": "Page Title",
//	    "lvl2": "Section"
//	  },
//	  "position": 2,
//	  "objectID": "https:--example.com-page#section-2"
//	}
package model
