// Package model defines shared crawler data types passed between pipeline stages.
//
// Record examples:
//
// Main page record:
//
//	{
//	  "url": "https://example.com/page",
//	  "type": "lvl1",
//	  "title": "Page Title",
//	  "description": "Page description",
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
//	  "type": "lvl2",
//	  "title": "Page Title",
//	  "description": "Page description",
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
//	  "type": "content",
//	  "title": "Page Title",
//	  "description": "Page description",
//	  "content": "Paragraph text",
//	  "hierarchy": {
//	    "lvl1": "Page Title",
//	    "lvl2": "Section"
//	  },
//	  "position": 2,
//	  "objectID": "https:--example.com-page#section-2"
//	}
package model
